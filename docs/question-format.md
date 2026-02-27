# Question Format Reference

This document describes the TOML file formats used for questions and question groups.

## Question files (`.toml`)

Each question is a single TOML file. The question's ID is the file path relative to the bank root, without the `.toml` extension (e.g., `vm/paging-001`).

### Fields

| Field | Required | Description |
|---|---|---|
| `stem` | yes | The question prompt. Supports Markdown and LaTeX. |
| `topic` | yes | Hierarchical category string, segments separated by `/` (e.g., `virtual-memory/paging`). |
| `type` | no | `"multiple-choice"`, `"true-false"`, `"short-answer"`, or `"fill-in-the-blank"`. Inferred from other fields if omitted. |
| `choices` | for multiple-choice | Array of answer choices. |
| `answer_tf` | for true-false | Boolean answer (`true` or `false`). Setting this also sets `type = "true-false"`. |
| `answer` | for short-answer | The correct answer string. Setting this also sets `type = "short-answer"`. |
| `answer_space` | no | Height of the answer blank box for short-answer questions (e.g. `"2in"`). Defaults to the `\defaultanswerlen` macro (`1in`; overridable per-exam via `\renewcommand` in the exam's `preamble`). |
| `blanks` | for fill-in-the-blank | Map of blank names to blank definitions. Setting this also sets `type = "fill-in-the-blank"`. Each blank name must appear in the stem as `[name]`. |
| `explanation` | no | Explanation of the correct answer, shown in solutions. |
| `difficulty` | no | `"easy"`, `"medium"`, or `"hard"`. |
| `tags` | no | Array of keyword strings for filtering. |
| `figure` | no | Path to a figure file to include alongside the stem. Image files (`.png`, `.jpg`, etc.) are included with `\includegraphics`; `.tex` files are included with `\input` (for TikZ figures). |
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

### Short-answer example

```toml
stem = """
What system call does a process use to create a child process?
"""
answer = "fork()"
explanation = """
fork() creates a new process by duplicating the calling process.
"""

topic = "processes"
difficulty = "easy"
tags = ["system-calls", "fork"]
```

### Fill-in-the-blank example (single blank)

```toml
stem = "The system call to create a child process is [blank1]."

topic = "processes"
difficulty = "easy"
tags = ["system-calls", "fork"]

[blanks.blank1]
answers = ["fork()", "fork"]
```

### Fill-in-the-blank example (multiple blanks)

```toml
stem = "A [lock_type] ensures only [n] thread(s) can be in a critical section."

topic = "concurrency"
difficulty = "medium"
tags = ["locks", "mutex"]

[blanks.lock_type]
answers = ["mutex"]
size = "1.5in"

[blanks.n]
answers = ["1", "one"]
```

Each blank has:

- `answers` (string array, required) — list of accepted answers (any match is correct)
- `size` (string, optional) — width of the underline in LaTeX output (e.g. `"1.5in"`). Defaults to `"1in"`.

Blank names in the stem use `[name]` syntax. In LaTeX output, blanks render as `\underline{\hspace{size}}`; in Canvas QTI, they map to `fill_in_multiple_blanks_question`.

**Important:** Because TOML table headers like `[blanks.blank1]` capture all subsequent key-value pairs, place all flat fields (`topic`, `difficulty`, `tags`, etc.) _before_ any `[blanks.*]` sections.

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
| `stem` | yes | Shared scenario or instructions shown above all parts. Use `GROUP:first` and `GROUP:last` in `\ref{}` as portable placeholders for the group's first/last part labels (e.g., `\ref{GROUP:first}`). |
| `topic` | yes | Inherited by parts that don't set their own `topic`. |
| `difficulty` | no | Inherited by parts that don't set their own `difficulty`. |
| `tags` | no | Inherited by parts that don't set their own `tags`. |
| `figure` | no | Image to show with the shared stem. |
| `parts` | yes | Array of part tables (at least one required). |

### Part fields

Each `[[parts]]` entry accepts the same fields as a standalone question (`stem`, `choices`, `answer_tf`, `answer`, `explanation`, `type`, `difficulty`, `tags`, `points`). Parts inherit `topic`, `difficulty`, and `tags` from the group when not explicitly set.

### Group example

```toml
stem = """
For questions \ref{GROUP:first}--\ref{GROUP:last},
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

## File naming conventions

- Questions: `<topic>-<NNN>.toml`, e.g. `vm-001.toml`
- Groups: `<topic>-<NNN>.group.toml`, e.g. `processes-group-001.group.toml`
- Subdirectories are allowed and become part of the question ID

## Question IDs

A question's ID is its path relative to the bank root, without the file extension:

- `vm-001.toml` → ID `vm-001`
- `virtual-memory/paging-001.toml` → ID `virtual-memory/paging-001`
- `processes-group-001.group.toml` → group ID `processes-group-001`, part IDs `processes-group-001/1`, `processes-group-001/2`, …
