package question

import (
	"bytes"

	"github.com/pelletier/go-toml/v2"
)

type questionTemplate struct {
	Stem        string           `toml:"stem,multiline" comment:"Required. The question prompt."`
	Topic       string           `toml:"topic" comment:"Used to organize questions."`
	Answer      *bool            `toml:"answer,omitempty" comment:"Correct answer (true/false)"`
	Explanation string           `toml:"explanation,multiline" comment:"Explanation of the answer for solutions."`
	Choices     []choiceTemplate `toml:"choices,omitempty" comment:"Answer choices"`
	Type        string           `toml:"type" comment:"Question type: 'multiple-choice' (default) or 'true-false'"`
	Difficulty  string           `toml:"difficulty"`
	Tags        []string         `toml:"tags" comment:"Keywords to categorize and find questions"`
	Figure      string           `toml:"figure,omitempty,commented" comment:"Optional figure path to include alongside the question stem"`
	Points      int              `toml:"points,omitempty,commented" comment:"Point value; treated as 1 if omitted"`
}

type choiceTemplate struct {
	Text    string `toml:"text"`
	Correct bool   `toml:"correct,omitempty"`
}

func encodeTemplate(tmpl questionTemplate) ([]byte, error) {
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	enc.SetArraysMultiline(true)
	if err := enc.Encode(tmpl); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MultipleChoiceTemplate returns TOML bytes for a multiple-choice question template.
func MultipleChoiceTemplate() ([]byte, error) {
	return encodeTemplate(questionTemplate{
		Type:        "multiple-choice",
		Difficulty:  "medium",
		Tags:        []string{},
		Stem:        "\n",
		Explanation: "\n",
		Points:      1,
		Choices: []choiceTemplate{
			{Text: "", Correct: true},
			{Text: ""},
			{Text: ""},
		},
	})
}

// TrueFalseTemplate returns TOML bytes for a true/false question template.
func TrueFalseTemplate() ([]byte, error) {
	f := false
	return encodeTemplate(questionTemplate{
		Type:        "true-false",
		Difficulty:  "medium",
		Tags:        []string{},
		Stem:        "\n",
		Explanation: "\n",
		Points:      1,
		Answer:      &f,
	})
}
