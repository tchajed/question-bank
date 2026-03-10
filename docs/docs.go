// Package docs provides embedded format reference documentation.
package docs

import (
	_ "embed"
	"strings"
)

// QuestionFormat is the question format reference documentation.
//
//go:embed question-format.md
var QuestionFormat string

// ExamFormat is the exam format reference documentation.
//
//go:embed exam-format.md
var ExamFormat string

// importPromptTemplate is the LLM import prompt template.
// It contains a {{QUESTION_FORMAT}} placeholder that gets replaced
// with the actual question format reference.
//
//go:embed import-prompt.md
var importPromptTemplate string

// ImportPrompt returns the full LLM prompt for importing exam content into
// question-bank TOML files. It fills in the question and exam format
// references.
func ImportPrompt() string {
	return strings.ReplaceAll(
		strings.ReplaceAll(importPromptTemplate,
			"{{QUESTION_FORMAT}}", QuestionFormat),
		"{{EXAM_FORMAT}}", ExamFormat)
}
