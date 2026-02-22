package question_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tchajed/question-bank/question"
)

func TestParseFile(t *testing.T) {
	q, err := question.ParseFile("../testdata/vm-001.toml")
	require.NoError(t, err)

	assert.Equal(t, "virtual-memory/paging", q.Topic)
	assert.Equal(t, question.DifficultyMedium, q.Difficulty)
	assert.Equal(t, question.MultipleChoice, q.Type)
	assert.Equal(t, "figures/two-level-page-table.svg", q.Figure)
	assert.Equal(t, 2, q.Points)
	assert.NotEmpty(t, q.Stem)
	require.Len(t, q.Choices, 3)

	var correct int
	for _, c := range q.Choices {
		if c.Correct {
			correct++
		}
	}
	assert.Equal(t, 1, correct, "exactly one choice should be correct")
}

func TestParseTrueFalse(t *testing.T) {
	q, err := question.ParseFile("../testdata/vm-002.toml")
	require.NoError(t, err)

	assert.Equal(t, question.TrueFalse, q.Type)
	require.NotNil(t, q.Answer)
	assert.False(t, *q.Answer)

	require.Len(t, q.Choices, 2)
	assert.Equal(t, question.Choice{Text: "True", Correct: false}, q.Choices[0])
	assert.Equal(t, question.Choice{Text: "False", Correct: true}, q.Choices[1])
}

func TestTrueFalseMissingAnswer(t *testing.T) {
	_, err := question.Parse([]byte(`topic = "os"
type = "true-false"
stem = "The sky is blue."`))
	assert.Error(t, err)
}

func TestParseRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid minimal",
			input: `topic = "os"
type = "multiple-choice"
stem = "What is 2+2?"`,
		},
		{
			name: "missing topic",
			input: `type = "multiple-choice"
stem = "What is 2+2?"`,
			wantErr: true,
		},
		{
			name: "missing stem",
			input: `topic = "os"
type = "multiple-choice"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := question.Parse([]byte(tt.input))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTypeDefault(t *testing.T) {
	t.Run("no type defaults to multiple-choice", func(t *testing.T) {
		q, err := question.Parse([]byte(`topic = "os"
stem = "What is 2+2?"`))
		require.NoError(t, err)
		assert.Equal(t, question.MultipleChoice, q.Type)
	})

	t.Run("answer set defaults to true-false", func(t *testing.T) {
		q, err := question.Parse([]byte(`topic = "os"
stem = "The sky is blue."
answer = true`))
		require.NoError(t, err)
		assert.Equal(t, question.TrueFalse, q.Type)
		require.NotNil(t, q.Answer)
		assert.True(t, *q.Answer)
	})
}
