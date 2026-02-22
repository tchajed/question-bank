// Package exam provides a format for grouping questions into exams.
//
// We use the term exam but this equally applies to exams, quizzes, midterms,
// homework, and practice exams.
package exam

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"github.com/tchajed/question-bank/question"
)

type Section struct {
	// Name is a header for the section
	Name string `toml:"name"`
	// Questions in the section (given by identifier)
	Questions []string `toml:"questions"`
}

// Exam represents a collection of questions along with exam metadata.
type Exam struct {
	// CourseCode is the course identifier, e.g. "CS 537".
	CourseCode string `toml:"course_code,omitempty"`
	// Title is the exam name, e.g. "Midterm 1".
	Title string `toml:"title,omitempty"`
	// Semester is the term, e.g. "SP 26".
	Semester string `toml:"semester,omitempty"`
	// Duration is the time allowed, e.g. "75 minutes".
	Duration string `toml:"duration,omitempty"`
	// CoverPage is freeform LaTeX for the body of the coverpages environment.
	// It may reference \ExamCourse, \ExamTitle, \ExamSemester, \ExamDuration,
	// and \ExamNumQuestions macros which are defined by the template.
	CoverPage string `toml:"cover_page,omitempty"`
	// Preamble is optional extra LaTeX inserted after the standard \usepackage lines.
	Preamble string `toml:"preamble,omitempty"`

	Sections []Section `toml:"sections"`
}

// merge returns a new Exam where non-zero fields from override replace those in base.
func merge(base, override *Exam) *Exam {
	result := *base
	if override.CourseCode != "" {
		result.CourseCode = override.CourseCode
	}
	if override.Title != "" {
		result.Title = override.Title
	}
	if override.Semester != "" {
		result.Semester = override.Semester
	}
	if override.Duration != "" {
		result.Duration = override.Duration
	}
	if override.CoverPage != "" {
		result.CoverPage = override.CoverPage
	}
	if override.Preamble != "" {
		result.Preamble = override.Preamble
	}
	if len(override.Sections) > 0 {
		result.Sections = override.Sections
	}
	return &result
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

// LoadWithDefaults reads an exam TOML file and merges it with a defaults.toml
// file in the same directory (if present). Fields in the exam file take
// precedence over defaults, so defaults.toml is a good place for course-level
// settings like CourseCode and CoverPage.
func LoadWithDefaults(path string) (*Exam, error) {
	dir := filepath.Dir(path)
	defaultsPath := filepath.Join(dir, "defaults.toml")

	var defaults Exam
	defaultsData, err := os.ReadFile(defaultsPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("reading defaults: %w", err)
	}
	if err == nil {
		dec := toml.NewDecoder(bytes.NewReader(defaultsData))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&defaults); err != nil {
			return nil, fmt.Errorf("parsing defaults.toml: %w", err)
		}
	}

	e, err := ParseFile(path)
	if err != nil {
		return nil, err
	}
	return merge(&defaults, e), nil
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
