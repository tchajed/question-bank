// Package docs provides embedded format reference documentation.
package docs

import _ "embed"

// QuestionFormat is the question format reference documentation.
//
//go:embed question-format.md
var QuestionFormat string

// ExamFormat is the exam format reference documentation.
//
//go:embed exam-format.md
var ExamFormat string
