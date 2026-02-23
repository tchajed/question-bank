# Exam Format Reference

This document describes the TOML file format used for exams.

## Exam files (`.toml`)

An exam file references questions from the bank by ID and organizes them into sections.

### Fields

| Field | Required | Description |
|---|---|---|
| `title` | no | Exam name, e.g. `"Midterm 1"`. |
| `course_code` | no | Course identifier, e.g. `"CS 537"`. |
| `semester` | no | Term, e.g. `"Spring 2026"`. |
| `cover_page` | no | Freeform LaTeX for the cover page body. May use macros `\ExamCourse`, `\ExamTitle`, `\ExamSemester`, `\ExamNumQuestions`. |
| `preamble` | no | Extra LaTeX inserted after standard `\usepackage` lines. |
| `sections` | yes | Array of section tables. |

### Section fields

Each `[[sections]]` entry has:

- `name` (string) — section header
- `questions` (array of strings) — question IDs in order

To include parts of a group, list the part IDs (e.g., `"processes-group-001/1"`, `"processes-group-001/2"`). Consecutive IDs from the same group are automatically rendered together under the shared group stem.

### Exam example

```toml
title = "Midterm 1"
semester = "Spring 2026"

[[sections]]
name = "Operating Systems"
questions = ["os-001", "processes-group-001/1", "processes-group-001/2"]

[[sections]]
name = "Virtual Memory"
questions = ["vm-001", "vm-002"]
```

### defaults.toml

A `defaults.toml` file placed in the same directory as exam files is automatically loaded and merged with each exam. Fields in the exam file take precedence. This is useful for course-level settings (`course_code`, `cover_page`) shared across multiple exams.
