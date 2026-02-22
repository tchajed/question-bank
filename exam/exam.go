// Package exam provides a format for grouping questions into exams.
//
// We use the term exam but this equally applies to quizzes, midterms, and practice exams.
package exam

type Section struct {
	// Name is a header for the section
	Name string `toml:"name"`
	// Questions in the section (given by identifier)
	Questions []string `toml:"questions"`
}

// Exam represents a collection of questions.
type Exam struct {
	// Sections in the exam.
	Sections []Section `toml:"sections"`
}
