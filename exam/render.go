package exam

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/tchajed/question-bank/question"
)

//go:embed exam.tmpl
var examTemplate string

// sectionItem is implemented by renderQuestion and renderGroup. Each produces
// its own LaTeX output via renderTeX.
type sectionItem interface {
	renderTeX() string
}

type renderSection struct {
	Name  string
	Items []sectionItem
}

type renderQuestion struct {
	Id           string
	Topic        string
	Difficulty   string
	Points       int
	Stem         string
	Type         string
	Choices      []question.Choice
	Answer       string // correct answer for short-answer questions
	AnswerSpace  string // box size override for short-answer (e.g. "2in"); empty means \defaultanswerlen
	// BlankStem and AnswerStem are pre-rendered stems for fill-in-the-blank
	// questions, with [name] placeholders replaced by underlines or answers.
	BlankStem  string
	AnswerStem string
	Explanation  string
	Figure       string // path to figure file (with extension)
	ShowMetadata bool
	Labels       []string // \label{} names attached to this \question
}

// isStandaloneTexFile reports whether the .tex file at path uses
// \documentclass{standalone} (possibly with options).
func isStandaloneTexFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, 128)
	n, _ := f.Read(buf)
	first := buf[:n]
	return bytes.Contains(first, []byte(`\documentclass`)) &&
		bytes.Contains(first, []byte(`{standalone}`))
}

// writeFigure writes a figure block to sb if figure is non-empty.
// Standalone .tex figures use \includestandalone; other .tex files use \input;
// all other files use \includegraphics.
func writeFigure(sb *strings.Builder, figure string) {
	if figure == "" {
		return
	}
	sb.WriteString("\n\\begin{center}\n")
	if strings.HasSuffix(figure, ".tex") {
		if isStandaloneTexFile(figure) {
			fmt.Fprintf(sb, "  \\includestandalone{%s}\n", strings.TrimSuffix(figure, ".tex"))
		} else {
			fmt.Fprintf(sb, "  \\input{%s}\n", figure)
		}
	} else {
		fmt.Fprintf(sb, "  \\includegraphics[width=0.5\\textwidth]{%s}\n", figure)
	}
	sb.WriteString("\\end{center}\n")
}

func (q *renderQuestion) renderTeX() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "%% %s | topic: %s | difficulty: %s", q.Id, q.Topic, q.Difficulty)
	if q.Points > 1 {
		fmt.Fprintf(&sb, " | points: %d", q.Points)
	}
	sb.WriteString("\n")

	fmt.Fprintf(&sb, "\\question[%d]\n", q.Points)
	for _, label := range q.Labels {
		fmt.Fprintf(&sb, "\\label{%s}\n", label)
	}
	if q.ShowMetadata {
		fmt.Fprintf(&sb, "{\\footnotesize\\textsf{%s \\textbar{} topic: %s \\textbar{} difficulty: %s}}\\\\[2pt]\n", q.Id, q.Topic, q.Difficulty)
	}
	if q.Type == string(question.FillInTheBlank) {
		sb.WriteString("\\ifprintanswers\n")
		sb.WriteString(q.AnswerStem)
		sb.WriteString("\n")
		if q.Explanation != "" {
			fmt.Fprintf(&sb, "\\paragraph{Solution:}%s\n", q.Explanation)
		}
		sb.WriteString("\\else\n")
		sb.WriteString(q.BlankStem)
		sb.WriteString("\n")
		sb.WriteString("\\fi\n")
	} else {
		sb.WriteString(q.Stem)
		sb.WriteString("\n")
	}

	writeFigure(&sb, q.Figure)

	if q.Type == string(question.FillInTheBlank) {
		// Already handled above; no choices to render.
	} else if q.Type == string(question.ShortAnswer) {
		sb.WriteString("\\ifprintanswers\n")
		fmt.Fprintf(&sb, "\\paragraph{Answer:}\\fbox{%s}\n", q.Answer)
		if q.Explanation != "" {
			fmt.Fprintf(&sb, "\\paragraph{Solution:}%s\n", q.Explanation)
		}
		sb.WriteString("\\else\n")
		answerLen := q.AnswerSpace
		if answerLen == "" {
			answerLen = `\defaultanswerlen`
		}
		fmt.Fprintf(&sb, "\\makeemptybox{%s}\n", answerLen)
		sb.WriteString("\\fi\n")
	} else {
		choicesEnv := "choices"
		if q.Type == string(question.TrueFalse) {
			choicesEnv = "checkboxes"
		}
		fmt.Fprintf(&sb, "\\begin{%s}\n", choicesEnv)
		for _, c := range q.Choices {
			if c.Correct {
				fmt.Fprintf(&sb, "  \\CorrectChoice %s\n", c.Text)
			} else {
				fmt.Fprintf(&sb, "  \\choice %s\n", c.Text)
			}
		}
		fmt.Fprintf(&sb, "\\end{%s}\n", choicesEnv)

		if q.Explanation != "" {
			sb.WriteString("\\ifprintanswers\n")
			fmt.Fprintf(&sb, "\\textbf{Solution:} %s\n", q.Explanation)
			sb.WriteString("\\fi\n")
		}
	}

	return sb.String()
}

type renderGroup struct {
	Id           string
	Topic        string
	Difficulty   string
	Stem         string
	Figure       string
	ShowMetadata bool
	Parts        []*renderQuestion
}

func (g *renderGroup) renderTeX() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "%% %s | topic: %s | difficulty: %s\n", g.Id, g.Topic, g.Difficulty)
	sb.WriteString("\\uplevel{\\vspace{1em}}\n")

	// Use EnvUplevel (environment form) instead of \uplevel{} so that verbatim
	// blocks in the stem are allowed.
	sb.WriteString("\\begin{EnvUplevel}\n")
	if g.ShowMetadata {
		fmt.Fprintf(&sb, "{\\footnotesize\\textsf{%s \\textbar{} topic: %s \\textbar{} difficulty: %s}}\\\\[2pt]\n",
			g.Id, g.Topic, g.Difficulty)
	}
	sb.WriteString(g.Stem)
	sb.WriteString("\n")
	writeFigure(&sb, g.Figure)
	sb.WriteString("\\end{EnvUplevel}\n")

	for _, part := range g.Parts {
		sb.WriteString(part.renderTeX())
		sb.WriteString("\n")
	}
	sb.WriteString("\\uplevel{\\vspace{1em}}\n")

	return sb.String()
}

// RenderData is the top-level data passed to the LaTeX template.
type RenderData struct {
	CourseCode   string
	Title        string
	Semester     string
	CoverPage    string
	Preamble     string
	NumQuestions int
	Sections     []renderSection
	PrintAnswers bool
}

// RenderOptions controls optional rendering behavior.
type RenderOptions struct {
	// PrintAnswers adds \printanswers to the document, revealing solutions.
	PrintAnswers bool
	// ShowMetadata renders question metadata (ID, topic, difficulty) as
	// visible text before each question.
	ShowMetadata bool
}

// figurePath prepends bankDir to figure. Returns "" if figure is "".
func figurePath(figure, bankDir string) string {
	if figure == "" {
		return ""
	}
	return filepath.Join(bankDir, figure)
}

// replaceGroupRefs replaces GROUP:first and GROUP:last placeholders with the
// group's actual label names (groupId+":first" and groupId+":last"). This lets
// group markdown use symbolic names instead of repeating the group ID. If
// groupId is empty the string is returned unchanged.
func replaceGroupRefs(s, groupId string) string {
	if groupId == "" {
		return s
	}
	s = strings.ReplaceAll(s, "GROUP:first", groupId+":first")
	s = strings.ReplaceAll(s, "GROUP:last", groupId+":last")
	return s
}

// buildRenderQuestion converts a question.Question to a renderQuestion.
// bankDir is prepended to any figure path. showMetadata controls the visible
// metadata annotation. groupId, when non-empty, enables GROUP:first/GROUP:last
// placeholder substitution in the question's markdown fields.
func buildRenderQuestion(q *question.Question, bankDir string, showMetadata bool, groupId string) *renderQuestion {
	choices := make([]question.Choice, len(q.Choices))
	for i, c := range q.Choices {
		choices[i] = question.Choice{
			Text:    markdownToTeX(replaceGroupRefs(c.Text, groupId)),
			Correct: c.Correct,
		}
	}
	stem := replaceGroupRefs(q.Stem, groupId)
	var blankStem, answerStem string
	if q.Type == question.FillInTheBlank {
		// Replace [name] placeholders before markdown conversion so that
		// blank names with underscores (e.g. lock_type) are handled correctly.
		bs := stem
		as := stem
		for name, b := range q.Blanks {
			as = strings.ReplaceAll(as, "["+name+"]", fmt.Sprintf("\\fbox{%s}", b.Answers[0]))
			bs = strings.ReplaceAll(bs, "["+name+"]", fmt.Sprintf("\\underline{\\hspace{%s}}", b.Size))
		}
		blankStem = markdownToTeX(bs)
		answerStem = markdownToTeX(as)
	}

	return &renderQuestion{
		Id:           q.Id,
		Topic:        q.Topic,
		Difficulty:   string(q.Difficulty),
		Points:       q.Points,
		Stem:         markdownToTeX(stem),
		Type:         string(q.Type),
		Choices:      choices,
		Answer:       markdownToTeX(replaceGroupRefs(q.Answer, groupId)),
		AnswerSpace:  q.AnswerSpace,
		BlankStem:    blankStem,
		AnswerStem:   answerStem,
		Explanation:  markdownToTeX(replaceGroupRefs(q.Explanation, groupId)),
		Figure:       figurePath(q.Figure, bankDir),
		ShowMetadata: showMetadata,
	}
}

// Render generates a LaTeX document for the exam. bankDir is used to compute
// absolute figure paths.
func (e *Exam) Render(resolved *ResolvedExam, bankDir string, opts RenderOptions) ([]byte, error) {
	numQuestions := 0
	sections := make([]renderSection, len(resolved.Sections))
	for i, sec := range resolved.Sections {
		var items []sectionItem
		for _, item := range sec.Items {
			switch v := item.(type) {
			case *question.Question:
				items = append(items, buildRenderQuestion(v, bankDir, opts.ShowMetadata, ""))
				numQuestions++
			case *question.QuestionGroup:
				rg := &renderGroup{
					Id:           v.Id,
					Topic:        v.Topic,
					Difficulty:   string(v.Difficulty),
					Stem:         markdownToTeX(replaceGroupRefs(v.Stem, v.Id)),
					Figure:       figurePath(v.Figure, bankDir),
					ShowMetadata: opts.ShowMetadata,
					Parts:        make([]*renderQuestion, len(v.Parts)),
				}
				for j, part := range v.Parts {
					rq := buildRenderQuestion(part, bankDir, opts.ShowMetadata, v.Id)
					if j == 0 {
						rq.Labels = append(rq.Labels, v.Id+":first")
					}
					if j == len(v.Parts)-1 {
						rq.Labels = append(rq.Labels, v.Id+":last")
					}
					rg.Parts[j] = rq
					numQuestions++
				}
				items = append(items, rg)
			}
		}
		sections[i] = renderSection{Name: sec.Name, Items: items}
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
		CoverPage:    coverPage,
		Preamble:     strings.TrimSpace(e.Preamble),
		NumQuestions: numQuestions,
		Sections:     sections,
		PrintAnswers: opts.PrintAnswers,
	}

	funcs := template.FuncMap{
		"renderItem": func(item sectionItem) (string, error) {
			return item.renderTeX(), nil
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
