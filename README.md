# question bank

[![CI](https://github.com/tchajed/question-bank/actions/workflows/ci.yml/badge.svg)](https://github.com/tchajed/question-bank/actions/workflows/ci.yml)

Write exam and quiz questions in a structured "question bank" format, then convert them to an exam or solutions sheet.

Each question is a TOML file, a bank is a collection of TOML files, and each exam is a TOML file that references the questions in a bank. Exams are rendered to PDF by inserting the questions into a LaTeX template and compiling it.

Quick demo (with example questions):

```sh
go run . --bank testdata/bank render-bank
go run . --bank testdata/bank render testdata/exams/exam.toml
```

## Questions

Here's a quick example:

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

An exam has some metadata and questions, listed by IDs in the question bank. Defaults can be factored out to a separate file, for several exams in the same course (this is particularly useful for the LaTeX cover page, not shown in the example below).

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

## Future work

- [ ] Better harness for LLM importing of questions
