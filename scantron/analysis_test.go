package scantron

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssignQuintiles(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		expected []int
	}{
		{"10 students", 10, []int{0, 0, 1, 1, 2, 2, 3, 3, 4, 4}},
		{"5 students", 5, []int{0, 1, 2, 3, 4}},
		{"15 students", 15, []int{0, 0, 0, 1, 1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 4}},
		{"7 students", 7, []int{0, 0, 1, 2, 2, 3, 4}},
		{"0 students", 0, []int{}},
		{"1 student", 1, []int{0}},
		{"3 students", 3, []int{0, 1, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := assignQuintiles(tt.n)
			assert.Equal(t, tt.expected, result)

			// Verify all values are in [0, 4]
			for _, q := range result {
				assert.GreaterOrEqual(t, q, 0)
				assert.LessOrEqual(t, q, 4)
			}
		})
	}
}

func TestAssignQuintilesMonotonic(t *testing.T) {
	// For any n, quintile assignments should be non-decreasing.
	for n := 1; n <= 50; n++ {
		result := assignQuintiles(n)
		for i := 1; i < len(result); i++ {
			assert.GreaterOrEqual(t, result[i], result[i-1],
				"n=%d: quintile[%d]=%d < quintile[%d]=%d", n, i, result[i], i-1, result[i-1])
		}
	}
}

// makeStudents creates test students with given scores and responses.
func makeStudents(scores []float64, responses [][]int) []*StudentRecord {
	records := make([]*StudentRecord, len(scores))
	for i, score := range scores {
		records[i] = &StudentRecord{
			ID:         "S" + string(rune('0'+i)),
			TotalScore: score,
			Responses:  responses[i],
		}
	}
	return records
}

func TestItemAnalysisOverallPctCorrect(t *testing.T) {
	// 10 students, 1 question with 4 choices, answer is choice 1.
	// 7 students answer correctly, 3 answer incorrectly.
	key := AnswerKey{1}
	questionIDs := []string{"q1"}
	numChoices := []int{4}

	responses := [][]int{
		{1}, {1}, {1}, {1}, {1}, {1}, {1}, // 7 correct
		{2}, {3}, {4}, // 3 incorrect
	}
	scores := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	records := makeStudents(scores, responses)

	stats, err := ItemAnalysis(records, key, questionIDs, numChoices)
	require.NoError(t, err)
	require.Len(t, stats, 1)

	assert.Equal(t, 1, stats[0].QuestionNum)
	assert.Equal(t, "q1", stats[0].QuestionID)
	assert.Equal(t, 4, stats[0].NumChoices)
	assert.InDelta(t, 70.0, stats[0].OverallPctCorrect, 0.01)
}

func TestItemAnalysisQuintileStats(t *testing.T) {
	// 10 students, 1 question with 3 choices, answer is choice 2.
	// Bottom quintile (lowest 2 scores): both wrong.
	// Top quintile (highest 2 scores): both correct.
	key := AnswerKey{2}
	questionIDs := []string{"q1"}
	numChoices := []int{3}

	// Scores in non-sorted order to verify sorting works.
	records := []*StudentRecord{
		{ID: "S5", TotalScore: 50, Responses: []int{2}},   // Q3
		{ID: "S1", TotalScore: 10, Responses: []int{1}},   // Q1 (bottom)
		{ID: "S10", TotalScore: 100, Responses: []int{2}}, // Q5 (top)
		{ID: "S2", TotalScore: 20, Responses: []int{3}},   // Q1 (bottom)
		{ID: "S9", TotalScore: 90, Responses: []int{2}},   // Q5 (top)
		{ID: "S3", TotalScore: 30, Responses: []int{2}},   // Q2
		{ID: "S8", TotalScore: 80, Responses: []int{2}},   // Q4
		{ID: "S4", TotalScore: 40, Responses: []int{1}},   // Q2
		{ID: "S7", TotalScore: 70, Responses: []int{1}},   // Q4
		{ID: "S6", TotalScore: 60, Responses: []int{2}},   // Q3
	}

	stats, err := ItemAnalysis(records, key, questionIDs, numChoices)
	require.NoError(t, err)
	require.Len(t, stats, 1)

	s := stats[0]

	// Bottom quintile (scores 10, 20): both wrong.
	assert.Equal(t, 2, s.ByQuintile[0].N)
	assert.InDelta(t, 0.0, s.ByQuintile[0].PctCorrect, 0.01)

	// Top quintile (scores 90, 100): both correct.
	assert.Equal(t, 2, s.ByQuintile[4].N)
	assert.InDelta(t, 100.0, s.ByQuintile[4].PctCorrect, 0.01)
}

func TestItemAnalysisResponseDist(t *testing.T) {
	// 10 students, 1 question with 3 choices.
	// Bottom quintile (2 students): one picks choice 1, one picks choice 3.
	key := AnswerKey{2}
	questionIDs := []string{"q1"}
	numChoices := []int{3}

	records := []*StudentRecord{
		{ID: "S1", TotalScore: 10, Responses: []int{1}},   // Q1 (bottom)
		{ID: "S2", TotalScore: 20, Responses: []int{3}},   // Q1 (bottom)
		{ID: "S3", TotalScore: 30, Responses: []int{2}},   // Q2
		{ID: "S4", TotalScore: 40, Responses: []int{2}},   // Q2
		{ID: "S5", TotalScore: 50, Responses: []int{2}},   // Q3
		{ID: "S6", TotalScore: 60, Responses: []int{2}},   // Q3
		{ID: "S7", TotalScore: 70, Responses: []int{2}},   // Q4
		{ID: "S8", TotalScore: 80, Responses: []int{0}},   // Q4, no response
		{ID: "S9", TotalScore: 90, Responses: []int{2}},   // Q5 (top)
		{ID: "S10", TotalScore: 100, Responses: []int{2}}, // Q5 (top)
	}

	stats, err := ItemAnalysis(records, key, questionIDs, numChoices)
	require.NoError(t, err)

	s := stats[0]

	// Bottom quintile: 1 chose A (idx 0), 1 chose C (idx 2), 0 chose B, 0 no-response.
	// ResponseDist has 4 entries: [A, B, C, no-response]
	assert.Len(t, s.ByQuintile[0].ResponseDist, 4)
	assert.InDelta(t, 0.5, s.ByQuintile[0].ResponseDist[0], 0.01) // choice 1 (A)
	assert.InDelta(t, 0.0, s.ByQuintile[0].ResponseDist[1], 0.01) // choice 2 (B)
	assert.InDelta(t, 0.5, s.ByQuintile[0].ResponseDist[2], 0.01) // choice 3 (C)
	assert.InDelta(t, 0.0, s.ByQuintile[0].ResponseDist[3], 0.01) // no response

	// Q4 quintile (students S7, S8): one chose B, one no response.
	assert.InDelta(t, 0.5, s.ByQuintile[3].ResponseDist[1], 0.01) // choice 2 (B)
	assert.InDelta(t, 0.5, s.ByQuintile[3].ResponseDist[3], 0.01) // no response
}

func TestItemAnalysisMultipleQuestions(t *testing.T) {
	// 10 students, 3 questions.
	key := AnswerKey{1, 3, 2}
	questionIDs := []string{"q1", "q2", "q3"}
	numChoices := []int{4, 4, 3}

	records := []*StudentRecord{
		{ID: "S1", TotalScore: 10, Responses: []int{1, 1, 1}},
		{ID: "S2", TotalScore: 20, Responses: []int{2, 3, 2}},
		{ID: "S3", TotalScore: 30, Responses: []int{1, 3, 2}},
		{ID: "S4", TotalScore: 40, Responses: []int{1, 2, 3}},
		{ID: "S5", TotalScore: 50, Responses: []int{1, 3, 2}},
		{ID: "S6", TotalScore: 60, Responses: []int{1, 3, 2}},
		{ID: "S7", TotalScore: 70, Responses: []int{1, 3, 1}},
		{ID: "S8", TotalScore: 80, Responses: []int{1, 3, 2}},
		{ID: "S9", TotalScore: 90, Responses: []int{1, 3, 2}},
		{ID: "S10", TotalScore: 100, Responses: []int{1, 3, 2}},
	}

	stats, err := ItemAnalysis(records, key, questionIDs, numChoices)
	require.NoError(t, err)
	require.Len(t, stats, 3)

	// Q1: 9 out of 10 correct (S2 is wrong).
	assert.InDelta(t, 90.0, stats[0].OverallPctCorrect, 0.01)

	// Q2: 8 out of 10 correct (S1 and S4 are wrong).
	assert.InDelta(t, 80.0, stats[1].OverallPctCorrect, 0.01)

	// Q3: 7 out of 10 correct (S1, S4, S7 are wrong).
	assert.InDelta(t, 70.0, stats[2].OverallPctCorrect, 0.01)

	// Verify question metadata.
	assert.Equal(t, "q2", stats[1].QuestionID)
	assert.Equal(t, 2, stats[1].QuestionNum)
	assert.Equal(t, 4, stats[1].NumChoices)
	assert.Equal(t, 3, stats[2].NumChoices)
}

func TestItemAnalysisErrors(t *testing.T) {
	records := []*StudentRecord{
		{ID: "S1", TotalScore: 50, Responses: []int{1, 2}},
	}

	t.Run("mismatched questionIDs length", func(t *testing.T) {
		_, err := ItemAnalysis(records, AnswerKey{1, 2}, []string{"q1"}, []int{4, 4})
		assert.Error(t, err)
	})

	t.Run("mismatched numChoices length", func(t *testing.T) {
		_, err := ItemAnalysis(records, AnswerKey{1, 2}, []string{"q1", "q2"}, []int{4})
		assert.Error(t, err)
	})

	t.Run("mismatched responses length", func(t *testing.T) {
		_, err := ItemAnalysis(records, AnswerKey{1}, []string{"q1"}, []int{4})
		assert.Error(t, err)
	})

	t.Run("no records", func(t *testing.T) {
		_, err := ItemAnalysis([]*StudentRecord{}, AnswerKey{1}, []string{"q1"}, []int{4})
		assert.Error(t, err)
	})
}

func TestWriteAnalysisCSV(t *testing.T) {
	stats := []QuestionStats{
		{
			QuestionNum:       1,
			QuestionID:        "q1",
			NumChoices:        4,
			OverallPctCorrect: 75.0,
			ByQuintile: [5]QuintileStats{
				{N: 2, PctCorrect: 50.0, ResponseDist: []float64{0.5, 0.25, 0.25, 0, 0}},
				{N: 2, PctCorrect: 50.0, ResponseDist: []float64{0.5, 0.5, 0, 0, 0}},
				{N: 2, PctCorrect: 100.0, ResponseDist: []float64{1, 0, 0, 0, 0}},
				{N: 2, PctCorrect: 100.0, ResponseDist: []float64{1, 0, 0, 0, 0}},
				{N: 2, PctCorrect: 100.0, ResponseDist: []float64{1, 0, 0, 0, 0}},
			},
		},
	}

	var buf bytes.Buffer
	err := WriteAnalysisCSV(&buf, stats)
	require.NoError(t, err)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	require.Len(t, lines, 2) // header + 1 data row

	assert.Equal(t, "Question,ID,OverallPctCorrect,Q1_PctCorrect,Q2_PctCorrect,Q3_PctCorrect,Q4_PctCorrect,Q5_PctCorrect", lines[0])
	assert.Equal(t, "1,q1,75.0,50.0,50.0,100.0,100.0,100.0", lines[1])
}

func TestWriteAnalysisJSON(t *testing.T) {
	stats := []QuestionStats{
		{
			QuestionNum:       1,
			QuestionID:        "q1",
			NumChoices:        3,
			OverallPctCorrect: 80.0,
			ByQuintile: [5]QuintileStats{
				{N: 2, PctCorrect: 50.0, ResponseDist: []float64{0.5, 0.5, 0, 0}},
				{N: 2, PctCorrect: 50.0, ResponseDist: []float64{0.5, 0.5, 0, 0}},
				{N: 2, PctCorrect: 100.0, ResponseDist: []float64{0, 1, 0, 0}},
				{N: 2, PctCorrect: 100.0, ResponseDist: []float64{0, 1, 0, 0}},
				{N: 2, PctCorrect: 100.0, ResponseDist: []float64{0, 1, 0, 0}},
			},
		},
	}

	var buf bytes.Buffer
	err := WriteAnalysisJSON(&buf, stats)
	require.NoError(t, err)

	// Verify it's valid JSON and round-trips.
	var decoded []QuestionStats
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)
	require.Len(t, decoded, 1)
	assert.Equal(t, "q1", decoded[0].QuestionID)
	assert.InDelta(t, 80.0, decoded[0].OverallPctCorrect, 0.01)
	assert.Equal(t, 2, decoded[0].ByQuintile[0].N)
}
