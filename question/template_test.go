package question_test

import (
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/require"
	"github.com/tchajed/question-bank/question"
)

// parseTemplate fills in required fields on a template and parses it.
func parseTemplate(t *testing.T, data []byte) *question.Question {
	t.Helper()
	var m map[string]any
	require.NoError(t, toml.Unmarshal(data, &m))
	m["topic"] = "test/topic"
	m["stem"] = "Test question?"
	patched, err := toml.Marshal(m)
	require.NoError(t, err)
	q, err := question.Parse(patched)
	require.NoError(t, err)
	return q
}

func TestMultipleChoiceTemplateParses(t *testing.T) {
	data, err := question.MultipleChoiceTemplate()
	require.NoError(t, err)
	q := parseTemplate(t, data)
	require.Equal(t, question.MultipleChoice, q.Type)
}

func TestTrueFalseTemplateParses(t *testing.T) {
	data, err := question.TrueFalseTemplate()
	require.NoError(t, err)
	q := parseTemplate(t, data)
	require.Equal(t, question.TrueFalse, q.Type)
}
