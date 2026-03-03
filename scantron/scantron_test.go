package scantron

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tchajed/question-bank/exam"
	"github.com/tchajed/question-bank/question"
)

const sampleCSV = `"LastName","FirstName","MI","ID","SpecialCodes","TotalScore","TOTALPCT","_1","_2","_3"
"Smith","Alice","A","1001","A","80","80","1","2","3"
"Jones","Bob","B","1002","B","70","70","3","1","2"
"Lee","Carol","","1003","A","90","90","1","3","1"
`

func TestParseCSV(t *testing.T) {
	data, err := ParseCSV(strings.NewReader(sampleCSV))
	require.NoError(t, err)

	assert.Equal(t, 3, data.NumQuestions)
	assert.Len(t, data.Records, 3)

	alice := data.Records[0]
	assert.Equal(t, "Smith", alice.LastName)
	assert.Equal(t, "Alice", alice.FirstName)
	assert.Equal(t, "A", alice.MI)
	assert.Equal(t, "1001", alice.ID)
	assert.Equal(t, "A", alice.SpecialCodes)
	assert.Equal(t, 80.0, alice.TotalScore)
	assert.Equal(t, 80.0, alice.TotalPct)
	assert.Equal(t, []int{1, 2, 3}, alice.Responses)

	carol := data.Records[2]
	assert.Equal(t, "", carol.MI)
	assert.Equal(t, 90.0, carol.TotalScore)
}

func TestParseCSVBlankResponse(t *testing.T) {
	csv := `"LastName","FirstName","MI","ID","SpecialCodes","TotalScore","TOTALPCT","_1","_2"
"Doe","Jane","","2001","A","50","50","1",""
`
	data, err := ParseCSV(strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, []int{1, 0}, data.Records[0].Responses)
}

func TestWriteCSV(t *testing.T) {
	data := &ScantronData{
		NumQuestions: 2,
		Records: []*StudentRecord{
			{
				LastName: "Smith", FirstName: "Alice", MI: "A", ID: "1001",
				SpecialCodes: "A", TotalScore: 80, TotalPct: 80,
				Responses: []int{1, 3},
			},
		},
	}

	var buf bytes.Buffer
	err := WriteCSV(&buf, data)
	require.NoError(t, err)

	// Parse it back
	roundTrip, err := ParseCSV(strings.NewReader(buf.String()))
	require.NoError(t, err)
	assert.Equal(t, data.NumQuestions, roundTrip.NumQuestions)
	assert.Equal(t, data.Records[0].LastName, roundTrip.Records[0].LastName)
	assert.Equal(t, data.Records[0].Responses, roundTrip.Records[0].Responses)
}

func TestPermutationReorder(t *testing.T) {
	// Permutation: position 0→2, 1→0, 2→1
	// (version question 1 maps to canonical question 3, etc.)
	perm := Permutation{2, 0, 1}

	result, err := perm.Reorder([]int{5, 6, 7})
	require.NoError(t, err)
	// version[0]=5 goes to canonical[2], version[1]=6 goes to canonical[0], version[2]=7 goes to canonical[1]
	assert.Equal(t, []int{6, 7, 5}, result)
}

func TestIdentityPermutation(t *testing.T) {
	perm := Identity(3)
	result, err := perm.Reorder([]int{1, 2, 3})
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, result)
}

func TestPermutationLengthMismatch(t *testing.T) {
	perm := Permutation{1, 0}
	_, err := perm.Reorder([]int{1, 2, 3})
	assert.Error(t, err)
}

// makeResolvedExam builds a simple ResolvedExam from question IDs.
func makeResolvedExam(ids ...string) *exam.ResolvedExam {
	items := make([]question.BankItem, len(ids))
	for i, id := range ids {
		items[i] = &question.Question{
			Id:    id,
			Stem:  "stem",
			Topic: "test",
			Type:  question.MultipleChoice,
			Choices: []question.Choice{
				{Text: "A", Correct: true},
				{Text: "B"},
				{Text: "C"},
			},
		}
	}
	return &exam.ResolvedExam{
		Sections: []exam.ResolvedSection{
			{Name: "Test", Items: items},
		},
	}
}

func TestDerivePermutation(t *testing.T) {
	canonical := makeResolvedExam("q1", "q2", "q3")
	version := makeResolvedExam("q3", "q1", "q2")

	perm, err := DerivePermutation(canonical, version)
	require.NoError(t, err)

	// version[0]=q3 → canonical[2], version[1]=q1 → canonical[0], version[2]=q2 → canonical[1]
	assert.Equal(t, Permutation{2, 0, 1}, perm)

	// Verify: if a student answered [5, 6, 7] on the version,
	// reordering gives us canonical order [6, 7, 5]
	result, err := perm.Reorder([]int{5, 6, 7})
	require.NoError(t, err)
	assert.Equal(t, []int{6, 7, 5}, result)
}

func TestDerivePermutationMismatch(t *testing.T) {
	canonical := makeResolvedExam("q1", "q2")
	version := makeResolvedExam("q1", "q3")

	_, err := DerivePermutation(canonical, version)
	assert.Error(t, err)
}

func TestReorderAll(t *testing.T) {
	data := &ScantronData{
		NumQuestions: 3,
		Records: []*StudentRecord{
			{ID: "1", SpecialCodes: "A", Responses: []int{1, 2, 3}},
			{ID: "2", SpecialCodes: "B", Responses: []int{5, 6, 7}},
			{ID: "3", SpecialCodes: "A", Responses: []int{4, 4, 4}},
		},
	}

	versions := VersionMap{
		// B: position 0→2, 1→0, 2→1
		"B": Permutation{2, 0, 1},
	}

	result, err := ReorderAll(data, versions)
	require.NoError(t, err)

	// A students: identity (no permutation)
	assert.Equal(t, []int{1, 2, 3}, result.Records[0].Responses)
	assert.Equal(t, []int{4, 4, 4}, result.Records[2].Responses)

	// B student: reordered
	assert.Equal(t, []int{6, 7, 5}, result.Records[1].Responses)
}

func TestGrade(t *testing.T) {
	key := AnswerKey{1, 3, 2}
	record := &StudentRecord{
		Responses: []int{1, 2, 2}, // correct, wrong, correct
	}

	graded, err := Grade(record, key)
	require.NoError(t, err)
	assert.Equal(t, []bool{true, false, true}, graded.Correct)
	assert.Equal(t, 2, graded.NumCorrect)
	assert.Equal(t, 3, graded.NumTotal)
}

func TestGradeSkipsNonGradeable(t *testing.T) {
	key := AnswerKey{1, 0, 2} // question 2 is short-answer (not gradeable)
	record := &StudentRecord{
		Responses: []int{1, 3, 1},
	}

	graded, err := Grade(record, key)
	require.NoError(t, err)
	assert.Equal(t, 1, graded.NumCorrect) // only q1 is correct
	assert.Equal(t, 2, graded.NumTotal)   // q2 is skipped
}

func TestDeriveAnswerKey(t *testing.T) {
	resolved := makeResolvedExam("q1", "q2")
	key := DeriveAnswerKey(resolved)
	// All questions in makeResolvedExam have choice A (index 0) as correct → 1-based index 1
	assert.Equal(t, AnswerKey{1, 1}, key)
}
