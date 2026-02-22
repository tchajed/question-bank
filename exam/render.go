package exam

import (
	"bytes"
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/tchajed/question-bank/question"
)

//go:embed exam.tmpl
var examTemplate string

type renderSection struct {
	Name      string
	Questions []*renderQuestion
}

type renderQuestion struct {
	Id           string
	Topic        string
	Difficulty   string
	Points       int
	Stem         string
	Type         string
	Choices      []question.Choice
	Explanation  string
	Figure       string // relative path for \includegraphics (no extension)
	ShowMetadata bool
}

// RenderData is the top-level data passed to the LaTeX template.
type RenderData struct {
	CourseCode   string
	Title        string
	Semester     string
	Duration     string
	CoverPage    string
	Preamble     string
	NumQuestions int
	Sections     []renderSection
	PrintAnswers bool
}

func renderQuestionTeX(q *renderQuestion) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%% %s | topic: %s | difficulty: %s", q.Id, q.Topic, q.Difficulty))
	if q.Points > 1 {
		sb.WriteString(fmt.Sprintf(" | points: %d", q.Points))
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("\\question[%d]\n", q.Points))
	if q.ShowMetadata {
		sb.WriteString(fmt.Sprintf("{\\footnotesize\\textsf{%s \\textbar{} topic: %s \\textbar{} difficulty: %s}}\\\\[2pt]\n", q.Id, q.Topic, q.Difficulty))
	}
	sb.WriteString(q.Stem)
	sb.WriteString("\n")

	if q.Figure != "" {
		sb.WriteString("\n\\begin{center}\n")
		sb.WriteString(fmt.Sprintf("  \\includegraphics[width=0.5\\textwidth]{%s}\n", q.Figure))
		sb.WriteString("\\end{center}\n")
	}

	choicesEnv := "choices"
	if q.Type == string(question.TrueFalse) {
		choicesEnv = "checkboxes"
	}
	sb.WriteString(fmt.Sprintf("\\begin{%s}\n", choicesEnv))
	for _, c := range q.Choices {
		if c.Correct {
			sb.WriteString(fmt.Sprintf("  \\CorrectChoice %s\n", c.Text))
		} else {
			sb.WriteString(fmt.Sprintf("  \\choice %s\n", c.Text))
		}
	}
	sb.WriteString(fmt.Sprintf("\\end{%s}\n", choicesEnv))

	if q.Explanation != "" {
		sb.WriteString("\\ifprintanswers\n")
		sb.WriteString(fmt.Sprintf("\\textbf{Solution:} %s\n", q.Explanation))
		sb.WriteString("\\fi\n")
	}

	return sb.String()
}

// RenderOptions controls optional rendering behavior.
type RenderOptions struct {
	// PrintAnswers adds \printanswers to the document, revealing solutions.
	PrintAnswers bool
	// ShowMetadata renders question metadata (ID, topic, difficulty) as
	// visible text before each question.
	ShowMetadata bool
}

// Render generates a LaTeX document for the exam. bankDir is used to compute
// absolute figure paths.
func (e *Exam) Render(resolved *ResolvedExam, bankDir string, opts RenderOptions) ([]byte, error) {
	numQuestions := 0
	sections := make([]renderSection, len(resolved.Sections))
	for i, sec := range resolved.Sections {
		qs := make([]*renderQuestion, len(sec.Questions))
		for j, q := range sec.Questions {
			points := q.Points
			if points == 0 {
				points = 1
			}
			choices := make([]question.Choice, len(q.Choices))
			for i, c := range q.Choices {
				choices[i] = question.Choice{
					Text:    markdownToTeX(c.Text),
					Correct: c.Correct,
				}
			}
			rq := &renderQuestion{
				Id:           q.Id,
				Topic:        q.Topic,
				Difficulty:   string(q.Difficulty),
				Points:       points,
				Stem:         markdownToTeX(q.Stem),
				Type:         string(q.Type),
				Choices:      choices,
				Explanation:  markdownToTeX(q.Explanation),
				ShowMetadata: opts.ShowMetadata,
			}
			if q.Figure != "" {
				fig := strings.TrimSuffix(q.Figure, filepath.Ext(q.Figure))
				rq.Figure = filepath.Join(bankDir, fig)
			}
			qs[j] = rq
			numQuestions++
		}
		sections[i] = renderSection{Name: sec.Name, Questions: qs}
	}

	// Normalize CoverPage: ensure it ends with exactly one newline.
	coverPage := strings.TrimRight(e.CoverPage, " \t\n")
	if coverPage != "" {
		coverPage += "\n"
	}

	data := &RenderData{
		CourseCode:   e.CourseCode,
		Title:        e.Title,
		Semester:     e.Semester,
		Duration:     e.Duration,
		CoverPage:    coverPage,
		Preamble:     strings.TrimSpace(e.Preamble),
		NumQuestions: numQuestions,
		Sections:     sections,
		PrintAnswers: opts.PrintAnswers,
	}

	funcs := template.FuncMap{
		"renderQuestion": func(q *renderQuestion) (string, error) {
			return renderQuestionTeX(q), nil
		},
	}

	tmpl, err := template.New("exam").Delims("<<", ">>").Funcs(funcs).Parse(examTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}
	return buf.Bytes(), nil
}
