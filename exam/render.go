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

//go:embed student_header.tmpl
var studentHeaderTemplate string

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
	Id          string
	Topic       string
	Difficulty  string
	Points      int
	Stem        string
	Type        string
	Choices     []question.Choice
	Answer      string // correct answer for short-answer questions
	AnswerSpace string // box size override for short-answer (e.g. "2in"); empty means \defaultanswerlen
	// BlankStem and AnswerStem are pre-rendered stems for fill-in-the-blank
	// questions, with [name] placeholders replaced by underlines or answers.
	BlankStem    string
	AnswerStem   string
	Explanation  string
	Figure       string // path to figure file (with extension)
	ShowMetadata bool
	Labels       []string // \label{} names attached to this \question
	// StudentResponse is used in student sheet mode.
	// 0 = not in student mode; positive = 1-based student choice index;
	// -1 = student mode but no response given.
	StudentResponse int
}

// StudentResponse describes one student's answers for rendering a
// personalized exam sheet with color-coded feedback.
type StudentResponse struct {
	Name      string // "LastName, FirstName"
	ID        string
	Responses []int // 1-based answer per question (0 = no response)
	Earned    int   // points earned
	Total     int   // total points possible
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
		if q.StudentResponse != 0 {
			// Student sheet mode: color-code choices based on student's answer.
			correctIdx := 0 // 1-based index of the correct choice
			for i, c := range q.Choices {
				if c.Correct {
					correctIdx = i + 1
					break
				}
			}
			studentChoice := q.StudentResponse // -1 means no response
			if studentChoice == -1 {
				studentChoice = 0
			}
			for i, c := range q.Choices {
				idx := i + 1 // 1-based
				if idx == correctIdx && idx == studentChoice {
					// Student chose the correct answer
					fmt.Fprintf(&sb, "  \\choice \\correctmark{%s}\n", c.Text)
				} else if idx == correctIdx {
					// This is the correct answer (student chose wrong or no response)
					fmt.Fprintf(&sb, "  \\choice \\correctmark{%s}\n", c.Text)
				} else if idx == studentChoice {
					// Student chose this wrong answer
					fmt.Fprintf(&sb, "  \\choice \\wrongmark{%s}\n", c.Text)
				} else {
					fmt.Fprintf(&sb, "  \\choice %s\n", c.Text)
				}
			}
		} else {
			for _, c := range q.Choices {
				if c.Correct {
					fmt.Fprintf(&sb, "  \\CorrectChoice %s\n", c.Text)
				} else {
					fmt.Fprintf(&sb, "  \\choice %s\n", c.Text)
				}
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

// studentSheetPreamble is injected into the preamble for student feedback sheets.
// It defines xcolor-based commands for marking correct and wrong answers.
const studentSheetPreamble = `\usepackage{xcolor}
\newcommand{\correctmark}[1]{\textcolor{green!40!black}{\textbf{#1}}}
\newcommand{\wrongmark}[1]{\textcolor{red!50!black}{\textbf{#1}}}`

// RenderStudentSheet renders a personalized exam sheet for one student.
// The exam is rendered with the student's answers color-coded: correct
// choices in green, wrong choices in red, and the correct answer always
// shown in green.
func (e *Exam) RenderStudentSheet(resolved *ResolvedExam, bankDir string, student StudentResponse) ([]byte, error) {
	// Build a flat list of questions so we can index by position.
	flatQuestions := resolved.FlattenQuestions()

	if len(student.Responses) != len(flatQuestions) {
		return nil, fmt.Errorf("student %s has %d responses but exam has %d questions",
			student.ID, len(student.Responses), len(flatQuestions))
	}

	// Build the question index: maps question ID to flat position.
	questionIndex := make(map[string]int, len(flatQuestions))
	for i, q := range flatQuestions {
		questionIndex[q.Id] = i
	}

	numQuestions := 0
	sections := make([]renderSection, len(resolved.Sections))
	for i, sec := range resolved.Sections {
		var items []sectionItem
		for _, item := range sec.Items {
			switch v := item.(type) {
			case *question.Question:
				rq := buildRenderQuestion(v, bankDir, false, "")
				// Set student response for MC/TF questions.
				if idx, ok := questionIndex[v.Id]; ok {
					rq.StudentResponse = studentResponseValue(v, student.Responses[idx])
				}
				items = append(items, rq)
				numQuestions++
			case *question.QuestionGroup:
				rg := &renderGroup{
					Id:         v.Id,
					Topic:      v.Topic,
					Difficulty: string(v.Difficulty),
					Stem:       markdownToTeX(replaceGroupRefs(v.Stem, v.Id)),
					Figure:     figurePath(v.Figure, bankDir),
					Parts:      make([]*renderQuestion, len(v.Parts)),
				}
				for j, part := range v.Parts {
					rq := buildRenderQuestion(part, bankDir, false, v.Id)
					if j == 0 {
						rq.Labels = append(rq.Labels, v.Id+":first")
					}
					if j == len(v.Parts)-1 {
						rq.Labels = append(rq.Labels, v.Id+":last")
					}
					if idx, ok := questionIndex[part.Id]; ok {
						rq.StudentResponse = studentResponseValue(part, student.Responses[idx])
					}
					rg.Parts[j] = rq
					numQuestions++
				}
				items = append(items, rg)
			}
		}
		sections[i] = renderSection{Name: sec.Name, Items: items}
	}

	// Build the student header as cover page content.
	var pct float64
	if student.Total > 0 {
		pct = float64(student.Earned) / float64(student.Total) * 100
	}
	headerData := struct {
		Exam    *Exam
		Student StudentResponse
		Pct     float64
	}{e, student, pct}
	headerTmpl, err := template.New("student_header").Delims("<<", ">>").Parse(studentHeaderTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing student header template: %w", err)
	}
	var headerBuf bytes.Buffer
	if err := headerTmpl.Execute(&headerBuf, headerData); err != nil {
		return nil, fmt.Errorf("executing student header template: %w", err)
	}
	coverPage := headerBuf.String()

	// Combine the existing preamble with student sheet commands.
	preamble := strings.TrimSpace(e.Preamble)
	if preamble != "" {
		preamble += "\n"
	}
	preamble += studentSheetPreamble

	data := &RenderData{
		CourseCode:   e.CourseCode,
		Title:        e.Title,
		Semester:     e.Semester,
		CoverPage:    coverPage,
		Preamble:     preamble,
		NumQuestions: numQuestions,
		Sections:     sections,
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

// studentResponseValue returns the StudentResponse field value for a
// renderQuestion. For MC/TF questions, it converts 0 (no response) to -1
// and passes through positive values. For other question types (short-answer,
// fill-in-blank), it returns 0 (not in student mode) since they cannot be
// graded on a scantron.
func studentResponseValue(q *question.Question, response int) int {
	if q.Type != question.MultipleChoice && q.Type != question.TrueFalse {
		return 0
	}
	if response == 0 {
		return -1 // no response, but in student mode
	}
	return response
}
