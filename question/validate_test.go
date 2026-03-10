package question_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tchajed/question-bank/question"
)

func TestDeepValidateMultipleChoice(t *testing.T) {
	t.Run("no correct answer", func(t *testing.T) {
		q, err := question.Parse([]byte(`topic = "os"
stem = "What is 2+2?"
choices = [{text = "3"}, {text = "4"}, {text = "5"}]`))
		require.NoError(t, err)
		warnings := q.DeepValidate()
		require.Len(t, warnings, 1)
		assert.Contains(t, warnings[0], "no correct answer")
	})

	t.Run("multiple correct answers", func(t *testing.T) {
		q, err := question.Parse([]byte(`topic = "os"
stem = "What is 2+2?"
choices = [{text = "3", correct = true}, {text = "4", correct = true}]`))
		require.NoError(t, err)
		warnings := q.DeepValidate()
		require.Len(t, warnings, 1)
		assert.Contains(t, warnings[0], "2 correct answers")
	})

	t.Run("exactly one correct answer", func(t *testing.T) {
		q, err := question.Parse([]byte(`topic = "os"
stem = "What is 2+2?"
choices = [{text = "3"}, {text = "4", correct = true}]`))
		require.NoError(t, err)
		assert.Empty(t, q.DeepValidate())
	})
}

func TestDeepValidateNonMultipleChoice(t *testing.T) {
	q, err := question.Parse([]byte(`topic = "os"
stem = "The sky is blue."
answer_tf = true`))
	require.NoError(t, err)
	assert.Empty(t, q.DeepValidate())
}

func TestValidateBank(t *testing.T) {
	t.Run("no warnings for clean bank", func(t *testing.T) {
		bank, err := question.LoadBank("../testdata/bank")
		require.NoError(t, err)
		warnings := question.ValidateBank(bank)
		assert.Empty(t, warnings)
	})

	t.Run("warns on MC question with no correct answer", func(t *testing.T) {
		q, err := question.Parse([]byte(`topic = "os"
stem = "What is 2+2?"
choices = [{text = "3"}, {text = "4"}]`))
		require.NoError(t, err)
		q.Id = "test-001"

		bank := question.Bank{"test-001": q}
		warnings := question.ValidateBank(bank)
		require.Len(t, warnings, 1)
		assert.Equal(t, "test-001", warnings[0].Id)
		assert.Contains(t, warnings[0].Message, "no correct answer")
		assert.Contains(t, warnings[0].String(), "test-001:")
	})

	t.Run("warns on group parts with no correct answer", func(t *testing.T) {
		g, err := question.ParseGroup([]byte(`
stem = "A parent process calls fork()."
topic = "processes"

[[parts]]
stem = "Part one?"
choices = [{text = "A"}, {text = "B"}]

[[parts]]
stem = "Part two?"
choices = [{text = "X", correct = true}]
`))
		require.NoError(t, err)
		g.Id = "group-001"
		g.Parts[0].Id = "group-001/1"
		g.Parts[1].Id = "group-001/2"

		bank := question.Bank{g.Id: g}
		warnings := question.ValidateBank(bank)
		require.Len(t, warnings, 1)
		assert.Equal(t, "group-001/1", warnings[0].Id)
	})
}
