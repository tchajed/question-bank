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
	assert.Contains(t, out, "fork()")          // group stem
	assert.Contains(t, out, `\question[1]`)    // part 1
	assert.Contains(t, out, `\question[2]`)    // part 2 (points = 2)
	assert.Contains(t, out, `\CorrectChoice`)
	assert.Contains(t, out, `\ifprintanswers`) // explanations present
	// Group metadata in comment, not in visible text
	assert.Contains(t, out, "% processes-group-001")
	assert.NotContains(t, out, `\footnotesize`) // no metadata visible text by default
	// Labels for first and last question
	assert.Contains(t, out, `\label{processes-group-001:first}`)
	assert.Contains(t, out, `\label{processes-group-001:last}`)
	// GROUP:start/end placeholders are replaced with actual label names
	assert.Contains(t, out, `\ref{processes-group-001:first}`)
	assert.Contains(t, out, `\ref{processes-group-001:last}`)
	assert.NotContains(t, out, "GROUP:start")
	assert.NotContains(t, out, "GROUP:end")
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
