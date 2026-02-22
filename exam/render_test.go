package exam_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tchajed/question-bank/exam"
	"github.com/tchajed/question-bank/question"
)

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
