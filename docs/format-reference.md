# Question Bank Format Reference

This document describes the TOML file formats used for questions, question groups, and exams.

## Question files (`.toml`)

Each question is a single TOML file. The question's ID is the file path relative to the bank root, without the `.toml` extension (e.g., `vm/paging-001`).

### Fields

| Field | Required | Description |
|---|---|---|
| `stem` | yes | The question prompt. Supports Markdown and LaTeX. |
| `topic` | yes | Hierarchical category string, segments separated by `/` (e.g., `virtual-memory/paging`). |
| `type` | no | `"multiple-choice"` or `"true-false"`. Inferred from other fields if omitted. |
| `choices` | for multiple-choice | Array of answer choices. |
| `answer_tf` | for true-false | Boolean answer (`true` or `false`). Setting this also sets `type = "true-false"`. |
| `explanation` | no | Explanation of the correct answer, shown in solutions. |
| `difficulty` | no | `"easy"`, `"medium"`, or `"hard"`. |
| `tags` | no | Array of keyword strings for filtering. |
| `figure` | no | Path to an image file to include alongside the stem. |
| `points` | no | Point value. Defaults to `1`. |

### Choices

Each entry in `choices` is an inline table with:

- `text` (string) — the choice text
- `correct` (bool, optional) — marks the correct answer; defaults to `false`

Exactly one choice should be marked `correct = true`.

### Multiple-choice example

```toml
stem = """
A process uses a two-level page table. The page size is 4KB and each
entry is 4 bytes. How many bytes of page table memory are required in
the worst case for a process using 1GB of virtual address space?
"""
figure = "figures/two-level-page-table.png"
choices = [
  {text = "4MB", correct = true},
  {text = "8MB"},
  {text = "16MB"},
]
explanation = """
In the worst case all second-level tables are allocated...
"""

topic = "virtual-memory/paging"
difficulty = "medium"
tags = ["page-table", "address-translation"]
points = 2
```

### True-false example

```toml
stem = """
A TLB miss always requires accessing main memory.
"""
answer_tf = false
explanation = """
A TLB miss requires a page table walk, but the page table entries may
themselves be cached in the CPU cache.
"""

topic = "virtual-memory"
difficulty = "easy"
tags = ["tlb"]
```

### Stem formatting

The stem is rendered as Markdown/LaTeX. You can include:

- Markdown formatting: `**bold**`, `` `code` ``, etc.
- Fenced code blocks with language tags (e.g., ` ```c `)
- Markdown tables
- LaTeX math and macros

## Question group files (`.group.toml`)

A question group is a multi-part question with shared introductory text. The file suffix must be `.group.toml` (e.g., `processes-group-001.group.toml`).

The group's ID is the file path without `.group.toml`. Each part gets ID `group-id/N` (1-indexed), e.g., `processes-group-001/1`, `processes-group-001/2`.

### Group-level fields

| Field | Required | Description |
|---|---|---|
| `stem` | yes | Shared scenario or instructions shown above all parts. Supports LaTeX `\ref{}` to reference the first/last part labels. |
| `topic` | yes | Inherited by parts that don't set their own `topic`. |
| `difficulty` | no | Inherited by parts that don't set their own `difficulty`. |
| `tags` | no | Inherited by parts that don't set their own `tags`. |
| `figure` | no | Image to show with the shared stem. |
| `parts` | yes | Array of part tables (at least one required). |

### Part fields

Each `[[parts]]` entry accepts the same fields as a standalone question (`stem`, `choices`, `answer_tf`, `explanation`, `type`, `difficulty`, `tags`, `points`). Parts inherit `topic`, `difficulty`, and `tags` from the group when not explicitly set.

### Group example

```toml
stem = """
For questions \ref{processes-group-001:first}--\ref{processes-group-001:last},
consider this situation: A parent process calls fork(). Both the parent and
child run to completion without any other system calls.
"""

topic = "processes/fork"
difficulty = "medium"
tags = ["fork", "processes"]

[[parts]]
stem = "What does fork() return in the parent process?"
choices = [
  {text = "The PID of the child process", correct = true},
  {text = "0"},
  {text = "-1"},
]
explanation = "fork() returns the child's PID to the parent."

[[parts]]
stem = "What does fork() return in the child process?"
choices = [
  {text = "The PID of the parent process"},
  {text = "0", correct = true},
  {text = "The PID of the child process"},
]
explanation = "fork() returns 0 in the child."
points = 2
```

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

## File naming conventions

- Questions: `<topic>-<NNN>.toml`, e.g. `vm-001.toml`
- Groups: `<topic>-<NNN>.group.toml`, e.g. `processes-group-001.group.toml`
- Subdirectories are allowed and become part of the question ID

## Question IDs

A question's ID is its path relative to the bank root, without the file extension:

- `vm-001.toml` → ID `vm-001`
- `virtual-memory/paging-001.toml` → ID `virtual-memory/paging-001`
- `processes-group-001.group.toml` → group ID `processes-group-001`, part IDs `processes-group-001/1`, `processes-group-001/2`, …
