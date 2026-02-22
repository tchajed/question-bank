package exam_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tchajed/question-bank/exam"
	"github.com/tchajed/question-bank/question"
)

func TestParseFile(t *testing.T) {
	e, err := exam.ParseFile("../testdata/exams/exam.toml")
	require.NoError(t, err)

	require.Len(t, e.Sections, 2)
	assert.Equal(t, "Operating Systems", e.Sections[0].Name)
	assert.Equal(t, []string{"os-001"}, e.Sections[0].Questions)
	assert.Equal(t, "Virtual Memory", e.Sections[1].Name)
	assert.Equal(t, []string{"vm-001", "vm-002"}, e.Sections[1].Questions)
}

func TestResolve(t *testing.T) {
	bank, err := question.LoadBank("../testdata/bank")
	require.NoError(t, err)

	e, err := exam.ParseFile("../testdata/exams/exam.toml")
	require.NoError(t, err)

	resolved, err := e.Resolve(bank)
	require.NoError(t, err)

	require.Len(t, resolved.Sections, 2)
	require.Len(t, resolved.Sections[0].Questions, 1)
	assert.Equal(t, "os-001", resolved.Sections[0].Questions[0].Id)
	require.Len(t, resolved.Sections[1].Questions, 2)
	assert.Equal(t, "vm-001", resolved.Sections[1].Questions[0].Id)
	assert.Equal(t, "vm-002", resolved.Sections[1].Questions[1].Id)
}

func TestResolveMissingQuestion(t *testing.T) {
	e := &exam.Exam{
		Sections: []exam.Section{
			{Name: "Test", Questions: []string{"nonexistent"}},
		},
	}
	_, err := e.Resolve(map[string]*question.Question{})
	assert.ErrorContains(t, err, "nonexistent")
}

func TestLoadWithDefaults(t *testing.T) {
	e, err := exam.LoadWithDefaults("../testdata/exams/exam.toml")
	require.NoError(t, err)

	// Fields from exam.toml
	assert.Equal(t, "Midterm 1", e.Title)
	assert.Equal(t, "Spring 2026", e.Semester)
	assert.Equal(t, "75 minutes", e.Duration)

	// Fields from defaults.toml
	assert.Equal(t, "CS 537", e.CourseCode)
	assert.NotEmpty(t, e.CoverPage)

	// Sections from exam.toml
	require.Len(t, e.Sections, 2)
}
