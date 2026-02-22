// Package exam provides a format for grouping questions into exams.
//
// We use the term exam but this equally applies to exams, quizzes, midterms,
// homework, and practice exams.
package exam

import (
	"bytes"
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
	"github.com/tchajed/question-bank/question"
)

type Section struct {
	// Name is a header for the section
	Name string `toml:"name"`
	// Questions in the section (given by identifier)
	Questions []string `toml:"questions"`
}

// Exam represents a collection of questions.
type Exam struct {
	Sections []Section `toml:"sections"`
}

// ParseFile reads and parses an Exam from a TOML file.
func ParseFile(path string) (*Exam, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var e Exam
	dec := toml.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&e); err != nil {
		return nil, err
	}
	return &e, nil
}

// ResolvedSection is a section with full question contents resolved.
type ResolvedSection struct {
	Name      string
	Questions []*question.Question
}

// ResolvedExam is an exam with all question IDs resolved to their full contents.
type ResolvedExam struct {
	Sections []ResolvedSection
}

// Resolve resolves all question IDs in the exam using the provided bank,
// returning a ResolvedExam with full question contents. Returns an error if
// any question ID is not found in the bank.
func (e *Exam) Resolve(bank map[string]*question.Question) (*ResolvedExam, error) {
	resolved := &ResolvedExam{
		Sections: make([]ResolvedSection, len(e.Sections)),
	}
	for i, sec := range e.Sections {
		rs := ResolvedSection{
			Name:      sec.Name,
			Questions: make([]*question.Question, len(sec.Questions)),
		}
		for j, id := range sec.Questions {
			q, ok := bank[id]
			if !ok {
				return nil, fmt.Errorf("question %q not found in bank", id)
			}
			rs.Questions[j] = q
		}
		resolved.Sections[i] = rs
	}
	return resolved, nil
}
