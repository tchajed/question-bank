# question bank

[![CI](https://github.com/tchajed/question-bank/actions/workflows/ci.yml/badge.svg)](https://github.com/tchajed/question-bank/actions/workflows/ci.yml)

Write exam and quiz questions in a structured "question bank" format, then convert them to an exam or solutions sheet.

Each question is a TOML file, a bank is a collection of TOML files, and each exam is a TOML file that references the questions in a bank. Exams are rendered to PDF by inserting the questions into a LaTeX template and compiling it.

Quick demo (with example questions):

```sh
go run . --bank testdata/bank render-bank
go run . --bank testdata/bank render testdata/exams/exam.toml
```

## Getting started with a coding agent

This tool is designed to be easy to use with LLMs, which can quickly convert your existing questions to question-bank's toml files.

First install the binary:

```sh
go install github.com/tchajed/question-bank@latest
```

If you have a Google Doc, you can download it as markdown, which works quite reliably. There are various tools to convert a Word docx to markdown. You can also import a PDF, but it will take more time and tokens. Then, just use a version of this prompt in a coding agent (like Claude Code or Codex - you need it to be able to call the tool):

```txt
Run `question-bank docs --prompt` for instructions. Then import @midterm1.md to `./bank`.
```

You can then render the exam and double-check everything (the coding agent can figure out how to do this for you from the CLI `--help` output).

## Questions

Here's a quick example of what a question TOML file looks like:

```toml
stem = """
Which of these is **not** an application benefit of an operating system?
"""
choices = [
  {text = 'A set of simpler abstractions against which to program'},
  {text = 'Independence from specific hardware and devices'},
  {text = 'More control over how hardware is used', correct = true},
]
explanation = """
An operating system gives less direct control over the hardware to applications.
"""

topic = 'os'
difficulty = 'easy'
tags = []
```

Read the [question format reference](docs/question-reference.md) for a complete guide.

## Exams (and quizzes and homeworks)

An exam (or other assessment) has some metadata and questions, listed by IDs in the question bank. Defaults can be factored out to a separate file, for several exams in the same course (this is particularly useful for the LaTeX cover page, not shown in the example below).

```toml
course_code = "CS 537"
title = "Midterm 1"
semester = "Spring 2026"

[[sections]]
name = "Operating Systems"
questions = ["os-001", "processes-group-001/1", "processes-group-001/2"]

[[sections]]
name = "Virtual Memory"
questions = ["vm-001", "vm-002"]
```

Exams can be exported to a QTI zip file that can be imported into Canvas:

```sh
go run . --bank testdata/bank canvas tesdata/exams/exam.toml
```

## Canvas export

The `question-bank canvas` command can convert an exam file to a Canvas QTI zip file, which can then be imported into Canvas from Settings > Import course content.
