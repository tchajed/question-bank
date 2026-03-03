# Scantron Results Processing: Design Plan

## Overview

Three features for processing scantron CSV results from the testing center:

1. **Reorder** - Permute answer columns so all versions map to a canonical question order
2. **Sheets** - Render per-student exam sheets with color-coded answers (green=correct, red=incorrect)
3. **Item analysis** - Compute per-question statistics by score quintile

## Scantron CSV Format

```
"LastName","FirstName","MI","ID","SpecialCodes","TotalScore","TOTALPCT","_1","_2","_3",...
```

- `SpecialCodes` identifies the exam version (e.g., `"A"` or `"B"`)
- `_N` columns contain 1-based numeric answers (1=A, 2=B, 3=C, ...; 0 or blank = no response)

## Data Flow

```
scantron.csv ──┐
               ├─→ [Reorder] ──→ reordered.csv ──┬─→ [Sheets]   ──→ per-student PDFs
exam-v1.toml ──┤                                  │
exam-v2.toml ──┘                                  └─→ [Analysis] ──→ item analysis data
                                                  │
                                          exam.toml + bank
```

All three features share the same reordered data. The **reorder** step normalizes everything to a canonical question order, then **sheets** and **analysis** operate on that normalized form.

## New Package: `scantron`

A single new package handles CSV parsing, data types, permutation, and analysis.

### Data Types

```go
// StudentRecord holds one student's parsed scantron row.
type StudentRecord struct {
    LastName     string
    FirstName    string
    MI           string
    ID           string
    SpecialCodes string   // exam version identifier
    TotalScore   float64
    TotalPct     float64
    Responses    []int    // 1-based answer index per question; 0 = no response
}

// ScantronData holds all parsed records plus metadata.
type ScantronData struct {
    Records      []*StudentRecord
    NumQuestions  int
}
```

### CSV Parsing

`ParseCSV(r io.Reader) (*ScantronData, error)` - Reads the scantron CSV format. Detects the number of questions from the `_N` column headers. Returns structured data.

### Permutation

The permutation is derived by comparing two exam TOML files. Both reference the same questions from the bank, just in different orders. The resolved, flattened question lists (after expanding groups into individual parts) define the ordering.

```go
// Permutation maps positions: Permutation[i] = j means question at position i
// in the source ordering should move to position j in the target ordering.
type Permutation []int

// DerivePermutation compares two resolved exams and returns the permutation
// that maps version's question order to the canonical question order.
// Both exams must contain exactly the same set of question IDs.
func DerivePermutation(canonical, version *exam.ResolvedExam) (Permutation, error)
```

The derivation works by flattening each exam's sections into an ordered list of question IDs (expanding groups into individual parts), then computing the mapping. Returns an error if the question sets don't match.

```go
// Reorder applies a permutation to a student's responses, returning a new
// response slice in canonical order.
func (p Permutation) Reorder(responses []int) []int
```

### Version Mapping

A `VersionMap` associates `SpecialCodes` values with their permutations:

```go
type VersionMap map[string]Permutation // SpecialCodes value → permutation
```

The canonical version maps to the identity permutation.

### Answer Key

Generate the answer key from the canonical resolved exam + bank:

```go
// AnswerKey holds the correct 1-based answer index for each question.
// For MC/TF questions this is straightforward; short-answer/fill-in-blank
// questions are excluded from scantron grading (they don't appear on scantrons).
type AnswerKey []int

// DeriveAnswerKey extracts the correct answer index for each MC/TF question
// from a resolved exam.
func DeriveAnswerKey(resolved *exam.ResolvedExam) (AnswerKey, error)
```

This reuses the logic from `cmd/answer_key.go`'s `answerLetter()`, but returns numeric indices instead of letters. The shared logic (finding the correct choice index) should be extracted into a helper function in the `question` package:

```go
// question.CorrectChoiceIndex returns the 1-based index of the correct choice,
// or 0 if the question has no choices (short-answer, fill-in-blank).
func (q *Question) CorrectChoiceIndex() int
```

This method can then be used by both the existing `answer-key` command and the new scantron code.

### Reordered CSV Output

`WriteCSV(w io.Writer, data *ScantronData) error` - Writes the same CSV format as input, but with response columns reordered. This makes the output usable by external tools expecting the same format. After reordering, all students' responses are in canonical question order regardless of which version they took.

### Grading

```go
// GradedRecord augments a StudentRecord with per-question correctness.
type GradedRecord struct {
    Student    *StudentRecord
    Correct    []bool   // whether each response is correct
    NumCorrect int
    NumTotal   int
}

// Grade checks each response against the answer key.
func Grade(record *StudentRecord, key AnswerKey) *GradedRecord
```

### Item Analysis

```go
// QuestionStats holds per-question analysis data.
type QuestionStats struct {
    QuestionNum    int       // 1-based position
    QuestionID     string    // from the exam
    NumChoices     int
    OverallPctCorrect float64
    // ByQuintile[q] for quintile q (0=bottom, 4=top)
    ByQuintile     [5]QuintileStats
}

type QuintileStats struct {
    N             int       // number of students in this quintile
    PctCorrect    float64
    // ResponseDist[c] = fraction choosing choice c (0-indexed; index NumChoices = no response)
    ResponseDist  []float64
}

// ItemAnalysis computes per-question statistics grouped by score quintile.
// Students are sorted by TotalScore and divided into 5 equal groups.
func ItemAnalysis(records []*StudentRecord, key AnswerKey, resolved *exam.ResolvedExam) ([]QuestionStats, error)
```

Output formats:
- **CSV**: Two CSV files: (1) percent correct by quintile per question, (2) response distribution matrix by quintile per question
- **JSON**: A single JSON file with the full `[]QuestionStats` for programmatic use (and later report generation)

## Personalized Exam Sheets

### Approach

Extend the existing LaTeX rendering pipeline to support a "student feedback" mode. For each student, render the canonical exam with their responses marked:

- **Correct answer chosen**: choice highlighted in green
- **Wrong answer chosen**: their choice highlighted in red, correct answer highlighted in green
- **No response**: correct answer highlighted in green (so they can see what they missed)

### Implementation

Add a new render mode to the `exam` package:

```go
// StudentResponse describes one student's answers for rendering.
type StudentResponse struct {
    Name      string // "LastName, FirstName"
    ID        string
    Responses []int  // 1-based answer per question (0 = no response)
    Score     float64
}

// RenderStudentSheet renders a personalized exam sheet for one student.
// The exam is rendered with the student's answers color-coded.
func (e *Exam) RenderStudentSheet(
    resolved *ResolvedExam,
    bankDir string,
    student StudentResponse,
) ([]byte, error)
```

### LaTeX Changes

Add color support to the preamble (via `\usepackage{xcolor}`) and define custom commands:

```latex
\usepackage{xcolor}
\newcommand{\correctchoice}[1]{\colorbox{green!30}{#1}}
\newcommand{\wrongchoice}[1]{\colorbox{red!30}{#1}}
```

The `renderQuestion.renderTeX()` method needs a variant (or parameterization) that marks specific choices with these commands based on the student's response. Rather than modifying the existing method, add a new field to `renderQuestion`:

```go
type renderQuestion struct {
    // ... existing fields ...
    StudentResponse int  // 0 = not in student mode; 1-based = student's choice
}
```

When `StudentResponse > 0`, the rendering logic wraps the appropriate choice text:
- If `StudentResponse == correctIndex`: wrap in `\correctchoice{}`
- If `StudentResponse != correctIndex`: wrap student's choice in `\wrongchoice{}`, wrap correct choice in `\correctchoice{}`

The student sheet also includes a header with the student's name, ID, and score.

### Output

For a class of N students, produce either:
- N separate PDF files named `{ID}.pdf` in an output directory
- One merged PDF (using `pdfunite` or LaTeX `\includepdf`)

Start with separate files; merging can be added later.

## CLI Commands

Add a `scantron` command group with three subcommands:

### `question-bank scantron reorder`

```
question-bank scantron reorder [flags] <results.csv>

Flags:
  --canonical <exam.toml>        Exam file for the canonical question order
  --version <code>=<exam.toml>   Map a SpecialCodes value to its exam file (repeatable)
  -o, --output <file>            Output CSV path (default: stdout)
```

Example:
```bash
question-bank scantron reorder \
  --canonical exams/midterm1-v1.toml \
  --version B=exams/midterm1-v2.toml \
  results.csv -o reordered.csv
```

Students with `SpecialCodes` not matching any `--version` flag are assumed to be canonical order (identity permutation). So the canonical version's code doesn't need to be mapped explicitly.

### `question-bank scantron sheets`

```
question-bank scantron sheets [flags] <reordered.csv>

Flags:
  --exam <exam.toml>     Canonical exam file
  -o, --output-dir <dir> Output directory for PDFs (default: ./sheets)
```

Reads the reordered CSV, loads the exam + bank, renders one PDF per student.

### `question-bank scantron analysis`

```
question-bank scantron analysis [flags] <reordered.csv>

Flags:
  --exam <exam.toml>     Canonical exam file
  --format csv|json      Output format (default: csv)
  -o, --output <file>    Output path (default: stdout)
```

## Code Reuse Opportunities

1. **`question.CorrectChoiceIndex()`** - New method, used by both `cmd/answer_key.go` and `scantron.DeriveAnswerKey()`. The existing `answerLetter()` function in `cmd/answer_key.go` can be simplified to call this.

2. **Exam resolution and flattening** - `exam.Resolve()` already handles expanding groups. The permutation derivation uses this directly. Add a helper to flatten a `ResolvedExam` into an ordered `[]*question.Question` list since this is needed in multiple places (permutation, answer key, analysis).

3. **LaTeX rendering pipeline** - The student sheet rendering reuses `buildRenderQuestion()` and the template system. The main change is adding student-response awareness to choice rendering, which is localized to `renderQuestion.renderTeX()`.

4. **CSV writing** - The reordered CSV writer can be reused for any future CSV output needs.

## File Organization

```
scantron/
  scantron.go       # CSV parsing, StudentRecord, ScantronData
  permutation.go    # Permutation type, DerivePermutation, Reorder
  grade.go          # AnswerKey, GradedRecord, Grade
  analysis.go       # ItemAnalysis, QuestionStats, QuintileStats
  scantron_test.go  # Tests for all of the above

cmd/
  scantron.go       # Parent "scantron" command
  scantron_reorder.go
  scantron_sheets.go
  scantron_analysis.go
```

Changes to existing files:
- `question/question.go` - Add `CorrectChoiceIndex()` method
- `exam/exam.go` or `exam/resolve.go` - Add `FlattenQuestions()` helper
- `exam/render.go` - Add student response rendering mode
- `exam/exam.tmpl` - Add xcolor package and custom commands (conditionally, via preamble injection rather than changing the base template)
- `cmd/answer_key.go` - Refactor to use `CorrectChoiceIndex()`

## Testing Strategy

- **Unit tests** in `scantron/`: CSV parsing with sample data, permutation derivation from known exam orderings, answer key extraction, grading, quintile computation
- **Integration test**: End-to-end with a small exam (from `testdata/`) and a synthetic scantron CSV
- Add sample scantron CSV files to `testdata/`

## Future: Analysis Report

The item analysis JSON output is designed to be consumed by a future report generator. The report would:
- Render a per-question bar chart (% correct by quintile) - a good discriminating question shows increasing % correct across quintiles
- Render a response distribution heatmap per question
- Flag questions with poor discrimination (flat or inverted quintile curves)
- Flag distractors that no one chose

This can be a `question-bank scantron report` subcommand that reads the JSON output and produces a LaTeX/PDF report.
