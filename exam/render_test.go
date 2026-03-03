package exam_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tchajed/question-bank/exam"
	"github.com/tchajed/question-bank/question"
)

func TestRenderGroup(t *testing.T) {
	bank, err := question.LoadBank("../testdata/bank")
	require.NoError(t, err)

	e := &exam.Exam{
		Sections: []exam.Section{
			{
				Name:      "Processes",
				Questions: []string{"processes-group-001/1", "processes-group-001/2"},
			},
		},
	}

	resolved, err := e.Resolve(bank)
	require.NoError(t, err)

	bankDir, err := filepath.Abs("../testdata/bank")
	require.NoError(t, err)

	latex, err := e.Render(resolved, bankDir, exam.RenderOptions{})
	require.NoError(t, err)

	out := string(latex)
	assert.Contains(t, out, `\uplevel{`)
	assert.Contains(t, out, "fork()")       // group stem
	assert.Contains(t, out, `\question[1]`) // part 1
	assert.Contains(t, out, `\question[2]`) // part 2 (points = 2)
	assert.Contains(t, out, `\CorrectChoice`)
	assert.Contains(t, out, `\ifprintanswers`) // explanations present
	// Group metadata in comment, not in visible text
	assert.Contains(t, out, "% processes-group-001")
	assert.NotContains(t, out, `\footnotesize`) // no metadata visible text by default
	// Labels for first and last question
	assert.Contains(t, out, `\label{processes-group-001:first}`)
	assert.Contains(t, out, `\label{processes-group-001:last}`)
	// GROUP:first/end placeholders are replaced with actual label names
	assert.Contains(t, out, `\ref{processes-group-001:first}`)
	assert.Contains(t, out, `\ref{processes-group-001:last}`)
	assert.NotContains(t, out, "GROUP:first")
	assert.NotContains(t, out, "GROUP:last")
}

func TestRenderGroupPartSelection(t *testing.T) {
	bank, err := question.LoadBank("../testdata/bank")
	require.NoError(t, err)

	e := &exam.Exam{
		Sections: []exam.Section{
			{
				Name:      "Processes",
				Questions: []string{"processes-group-001/1"},
			},
		},
	}

	resolved, err := e.Resolve(bank)
	require.NoError(t, err)

	bankDir, err := filepath.Abs("../testdata/bank")
	require.NoError(t, err)

	latex, err := e.Render(resolved, bankDir, exam.RenderOptions{})
	require.NoError(t, err)

	out := string(latex)
	assert.Contains(t, out, `\uplevel{`)
	// Only part 1 (1 point); part 2 (2 points) excluded
	assert.Contains(t, out, `\question[1]`)
	assert.NotContains(t, out, `\question[2]`)
	// Single part gets both labels
	assert.Contains(t, out, `\label{processes-group-001:first}`)
	assert.Contains(t, out, `\label{processes-group-001:last}`)
}

func TestRenderGroupShowMetadata(t *testing.T) {
	bank, err := question.LoadBank("../testdata/bank")
	require.NoError(t, err)

	e := &exam.Exam{
		Sections: []exam.Section{
			{Name: "P", Questions: []string{"processes-group-001/1", "processes-group-001/2"}},
		},
	}
	resolved, err := e.Resolve(bank)
	require.NoError(t, err)

	bankDir, err := filepath.Abs("../testdata/bank")
	require.NoError(t, err)

	latex, err := e.Render(resolved, bankDir, exam.RenderOptions{ShowMetadata: true})
	require.NoError(t, err)

	out := string(latex)
	// Metadata appears in the uplevel block and on each part
	assert.Contains(t, out, `\footnotesize`)
	assert.Contains(t, out, "processes-group-001")
	assert.Contains(t, out, "processes-group-001/1")
	assert.Contains(t, out, "processes-group-001/2")
}

func TestRenderTikzFigure(t *testing.T) {
	bank, err := question.LoadBank("../testdata/bank")
	require.NoError(t, err)

	e := &exam.Exam{
		Sections: []exam.Section{
			{Name: "VM", Questions: []string{"vm-004"}},
		},
	}

	resolved, err := e.Resolve(bank)
	require.NoError(t, err)

	bankDir, err := filepath.Abs("../testdata/bank")
	require.NoError(t, err)

	latex, err := e.Render(resolved, bankDir, exam.RenderOptions{})
	require.NoError(t, err)

	out := string(latex)
	assert.Contains(t, out, `\includestandalone{`)
	assert.Contains(t, out, "figures/tlb-diagram")
	assert.NotContains(t, out, `\includegraphics`)
	assert.NotContains(t, out, `\input{`)
}

func TestRenderSmoke(t *testing.T) {
	bank, err := question.LoadBank("../testdata/bank")
	require.NoError(t, err)

	e, err := exam.LoadWithDefaults("../testdata/exams/exam.toml")
	require.NoError(t, err)

	resolved, err := e.Resolve(bank)
	require.NoError(t, err)

	bankDir, err := filepath.Abs("../testdata/bank")
	require.NoError(t, err)

	_, err = e.Render(resolved, bankDir, exam.RenderOptions{})
	require.NoError(t, err)
}

func TestRenderStudentSheetCorrectChoice(t *testing.T) {
	bank, err := question.LoadBank("../testdata/bank")
	require.NoError(t, err)

	// Use a simple MC question
	e := &exam.Exam{
		CourseCode: "CS 537",
		Title:      "Midterm 1",
		Semester:   "Spring 2026",
		Sections: []exam.Section{
			{Name: "OS", Questions: []string{"os-001"}},
		},
	}

	resolved, err := e.Resolve(bank)
	require.NoError(t, err)

	bankDir, err := filepath.Abs("../testdata/bank")
	require.NoError(t, err)

	// os-001 has 3 choices, correct is index 3 (1-based: the third choice)
	// Student chose the correct answer (3)
	student := exam.StudentResponse{
		Name:      "Smith, Alice",
		ID:        "1001",
		Responses: []int{3},
		Score:     100.0,
	}

	latex, err := e.RenderStudentSheet(resolved, bankDir, student)
	require.NoError(t, err)

	out := string(latex)
	// The correct choice should be marked with \correctchoice{}
	assert.Contains(t, out, `\correctchoice{`)
	// Should not have any \wrongchoice{}
	assert.NotContains(t, out, `\wrongchoice{`)
	// Student header should appear
	assert.Contains(t, out, "Smith, Alice")
	assert.Contains(t, out, "1001")
	assert.Contains(t, out, "100.0")
	// Preamble should include xcolor
	assert.Contains(t, out, `\usepackage{xcolor}`)
	assert.Contains(t, out, `\newcommand{\correctchoice}`)
	assert.Contains(t, out, `\newcommand{\wrongchoice}`)
	// Should NOT contain \CorrectChoice (exam class command) since we are in student mode
	assert.NotContains(t, out, `\CorrectChoice`)
}

func TestRenderStudentSheetWrongChoice(t *testing.T) {
	bank, err := question.LoadBank("../testdata/bank")
	require.NoError(t, err)

	e := &exam.Exam{
		Sections: []exam.Section{
			{Name: "OS", Questions: []string{"os-001"}},
		},
	}

	resolved, err := e.Resolve(bank)
	require.NoError(t, err)

	bankDir, err := filepath.Abs("../testdata/bank")
	require.NoError(t, err)

	// os-001 correct answer is choice 3 (1-based). Student chose 1 (wrong).
	student := exam.StudentResponse{
		Name:      "Jones, Bob",
		ID:        "1002",
		Responses: []int{1},
		Score:     0.0,
	}

	latex, err := e.RenderStudentSheet(resolved, bankDir, student)
	require.NoError(t, err)

	out := string(latex)
	// Should have both correct and wrong markings
	assert.Contains(t, out, `\correctchoice{`)
	assert.Contains(t, out, `\wrongchoice{`)
}

func TestRenderStudentSheetNoResponse(t *testing.T) {
	bank, err := question.LoadBank("../testdata/bank")
	require.NoError(t, err)

	e := &exam.Exam{
		Sections: []exam.Section{
			{Name: "OS", Questions: []string{"os-001"}},
		},
	}

	resolved, err := e.Resolve(bank)
	require.NoError(t, err)

	bankDir, err := filepath.Abs("../testdata/bank")
	require.NoError(t, err)

	// Student gave no response (0)
	student := exam.StudentResponse{
		Name:      "Lee, Carol",
		ID:        "1003",
		Responses: []int{0},
		Score:     0.0,
	}

	latex, err := e.RenderStudentSheet(resolved, bankDir, student)
	require.NoError(t, err)

	out := string(latex)
	// Correct answer should still be highlighted
	assert.Contains(t, out, `\correctchoice{`)
	// No wrong choice since no response was given
	assert.NotContains(t, out, `\wrongchoice{`)
}

func TestRenderStudentSheetShortAnswer(t *testing.T) {
	bank, err := question.LoadBank("../testdata/bank")
	require.NoError(t, err)

	// os-002 is a short-answer question
	e := &exam.Exam{
		Sections: []exam.Section{
			{Name: "Mixed", Questions: []string{"os-001", "os-002"}},
		},
	}

	resolved, err := e.Resolve(bank)
	require.NoError(t, err)

	bankDir, err := filepath.Abs("../testdata/bank")
	require.NoError(t, err)

	// Two questions: os-001 (MC) and os-002 (short-answer).
	// Short-answer questions should render normally (not in student mode).
	student := exam.StudentResponse{
		Name:      "Doe, Jane",
		ID:        "1004",
		Responses: []int{3, 0},
		Score:     50.0,
	}

	latex, err := e.RenderStudentSheet(resolved, bankDir, student)
	require.NoError(t, err)

	out := string(latex)
	// MC question should have student marking
	assert.Contains(t, out, `\correctchoice{`)
	// Short-answer question should render with its normal \makeemptybox
	assert.Contains(t, out, `\makeemptybox{`)
}

func TestRenderStudentSheetMultipleQuestions(t *testing.T) {
	bank, err := question.LoadBank("../testdata/bank")
	require.NoError(t, err)

	e := &exam.Exam{
		Sections: []exam.Section{
			{Name: "VM", Questions: []string{"vm-001", "vm-003"}},
		},
	}

	resolved, err := e.Resolve(bank)
	require.NoError(t, err)

	bankDir, err := filepath.Abs("../testdata/bank")
	require.NoError(t, err)

	// vm-001: 3 choices, correct is 1 (4MB)
	// vm-003: 4 choices, correct is 1
	// Student answers: vm-001=1 (correct), vm-003=2 (wrong)
	student := exam.StudentResponse{
		Name:      "Test, Student",
		ID:        "9999",
		Responses: []int{1, 2},
		Score:     50.0,
	}

	latex, err := e.RenderStudentSheet(resolved, bankDir, student)
	require.NoError(t, err)

	out := string(latex)
	// Both correct and wrong markings should appear (one of each from each question)
	assert.Contains(t, out, `\correctchoice{`)
	assert.Contains(t, out, `\wrongchoice{`)
	// Both sections should be rendered
	assert.Contains(t, out, `\section*{VM}`)
}

func TestRenderStudentSheetGroup(t *testing.T) {
	bank, err := question.LoadBank("../testdata/bank")
	require.NoError(t, err)

	e := &exam.Exam{
		Sections: []exam.Section{
			{Name: "Processes", Questions: []string{"processes-group-001/1", "processes-group-001/2"}},
		},
	}

	resolved, err := e.Resolve(bank)
	require.NoError(t, err)

	bankDir, err := filepath.Abs("../testdata/bank")
	require.NoError(t, err)

	// Two group parts, both are MC questions
	student := exam.StudentResponse{
		Name:      "Group, Test",
		ID:        "5555",
		Responses: []int{1, 1},
		Score:     50.0,
	}

	latex, err := e.RenderStudentSheet(resolved, bankDir, student)
	require.NoError(t, err)

	out := string(latex)
	// Group stem should be present
	assert.Contains(t, out, "fork()")
	// Student response marking should be present
	assert.Contains(t, out, `\correctchoice{`)
}

func TestRenderStudentSheetResponseMismatch(t *testing.T) {
	bank, err := question.LoadBank("../testdata/bank")
	require.NoError(t, err)

	e := &exam.Exam{
		Sections: []exam.Section{
			{Name: "OS", Questions: []string{"os-001"}},
		},
	}

	resolved, err := e.Resolve(bank)
	require.NoError(t, err)

	bankDir, err := filepath.Abs("../testdata/bank")
	require.NoError(t, err)

	// Wrong number of responses
	student := exam.StudentResponse{
		Name:      "Bad, Data",
		ID:        "0000",
		Responses: []int{1, 2, 3},
		Score:     0.0,
	}

	_, err = e.RenderStudentSheet(resolved, bankDir, student)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "3 responses but exam has 1 questions")
}

func TestRender(t *testing.T) {
	bank, err := question.LoadBank("../testdata/bank")
	require.NoError(t, err)

	e, err := exam.LoadWithDefaults("../testdata/exams/exam.toml")
	require.NoError(t, err)

	resolved, err := e.Resolve(bank)
	require.NoError(t, err)

	bankDir, err := filepath.Abs("../testdata/bank")
	require.NoError(t, err)

	latex, err := e.Render(resolved, bankDir, exam.RenderOptions{})
	require.NoError(t, err)

	out := string(latex)
	assert.Contains(t, out, `\documentclass[12pt,addpoints]{exam}`)
	assert.Contains(t, out, `\newcommand{\ExamCourse}{CS 537\xspace}`)
	assert.Contains(t, out, `\newcommand{\ExamTitle}{Midterm 1\xspace}`)
	assert.Contains(t, out, `\newcommand{\ExamSemester}{Spring 2026\xspace}`)
	assert.Contains(t, out, `\newcommand{\ExamNumQuestions}{\numquestions\xspace}`)
	assert.Contains(t, out, `\section*{Operating Systems}`)
	assert.Contains(t, out, `\section*{Virtual Memory}`)
	assert.Contains(t, out, `\question[1]`)
	assert.Contains(t, out, `\question[2]`)
	assert.Contains(t, out, `\CorrectChoice`)
	assert.Contains(t, out, `\includegraphics`)
	assert.Contains(t, out, `\ifprintanswers`)
	assert.Contains(t, out, `\begin{coverpages}`)
	assert.Contains(t, out, `\ExamCourse: \ExamTitle`)
}
