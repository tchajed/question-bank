package question_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tchajed/question-bank/question"
)

func TestParseGroup(t *testing.T) {
	data := []byte(`
stem = "A parent process calls fork()."
topic = "processes/fork"
difficulty = "medium"
tags = ["fork"]

[[parts]]
stem = "What does fork() return in the parent?"
choices = [{text = "Child PID", correct = true}, {text = "0"}]

[[parts]]
stem = "What does fork() return in the child?"
choices = [{text = "0", correct = true}, {text = "Parent PID"}]
points = 2
`)
	g, err := question.ParseGroup(data)
	require.NoError(t, err)

	assert.Equal(t, "A parent process calls fork().", g.Stem)
	assert.Equal(t, "processes/fork", g.Topic)
	assert.Equal(t, question.DifficultyMedium, g.Difficulty)
	assert.Equal(t, []string{"fork"}, g.Tags)
	require.Len(t, g.Parts, 2)

	// Parts inherit metadata from the group.
	assert.Equal(t, "processes/fork", g.Parts[0].Topic)
	assert.Equal(t, question.DifficultyMedium, g.Parts[0].Difficulty)
	assert.Equal(t, []string{"fork"}, g.Parts[0].Tags)
	assert.Equal(t, 1, g.Parts[0].Points) // defaults to 1
	assert.Equal(t, question.MultipleChoice, g.Parts[0].Type)

	assert.Equal(t, 2, g.Parts[1].Points)
}

func TestParseGroupFile(t *testing.T) {
	g, err := question.ParseGroupFile("../testdata/bank", "processes-group-001.group.toml")
	require.NoError(t, err)

	assert.Equal(t, "processes-group-001", g.Id)
	require.Len(t, g.Parts, 2)
	assert.Equal(t, "processes-group-001/1", g.Parts[0].Id)
	assert.Equal(t, "processes-group-001/2", g.Parts[1].Id)
}

func TestParseGroupRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "missing topic",
			input:   `stem = "foo"` + "\n[[parts]]\nstem = \"bar\"\nchoices = []",
			wantErr: "topic",
		},
		{
			name:    "missing stem",
			input:   "topic = \"os\"\n[[parts]]\nstem = \"bar\"\nchoices = []",
			wantErr: "stem",
		},
		{
			name:    "no parts",
			input:   `stem = "foo"` + "\ntopic = \"os\"",
			wantErr: "at least one part",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := question.ParseGroup([]byte(tt.input))
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestParseGroupPartInheritance(t *testing.T) {
	// A part that sets its own topic overrides the group's topic.
	data := []byte(`
stem = "Group stem"
topic = "group-topic"
difficulty = "easy"

[[parts]]
stem = "Part with own topic"
topic = "part-topic"
choices = [{text = "A", correct = true}]
`)
	g, err := question.ParseGroup(data)
	require.NoError(t, err)
	assert.Equal(t, "part-topic", g.Parts[0].Topic)
	assert.Equal(t, question.DifficultyEasy, g.Parts[0].Difficulty) // inherited
}

func TestLoadBankIncludesGroups(t *testing.T) {
	bank, err := question.LoadBank("../testdata/bank")
	require.NoError(t, err)

	assert.Contains(t, bank, "processes-group-001")
	assert.Contains(t, bank, "os-001")

	item := bank["processes-group-001"]
	g, ok := item.(*question.QuestionGroup)
	require.True(t, ok, "expected *QuestionGroup")
	assert.Equal(t, "processes-group-001", g.GetId())
	require.Len(t, g.Parts, 2)
	assert.Equal(t, "processes-group-001/1", g.Parts[0].Id)
}
