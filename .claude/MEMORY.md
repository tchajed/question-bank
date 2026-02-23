# question-bank project memory

## Architecture
- Go library (`github.com/tchajed/question-bank`) for managing exam questions as TOML files
- `question/` package: `Question`, `QuestionGroup`, `Bank`, parsing
- `exam/` package: `Exam`, `Section`, `Resolve`, `Render` (LaTeX via exam class)
- CLI via spf13/cobra in `cmd/`

## Question bank files
- Regular questions: `*.toml` → parsed as `*question.Question`
- Question groups: `*.group.toml` → parsed as `*question.QuestionGroup`
- `question.LoadBank(dir)` returns `question.Bank` (= `map[string]BankItem`)
- `BankItem` interface: `GetId() string` — implemented by `*Question` and `*QuestionGroup`

## QuestionGroup design
- `QuestionGroup.Parts []*Question`: reuses `Question` struct (no separate Part type)
- Parts inherit `Topic`, `Difficulty`, `Tags` from group if not set explicitly
- Part IDs: `"group-id/1"`, `"group-id/2"` (1-indexed), set in `ParseGroupFile`
- Group ID: file path with `.group.toml` stripped (e.g. `vm-group-001.group.toml` → `vm-group-001`)

## Exam TOML format
- `Section.Questions []SectionItem` — each item is an inline table:
  - Standalone: `{id = "os-001"}`
  - Whole group: `{group = "processes-group-001"}`
  - Subset of group: `{group = "processes-group-001", parts = [1, 3]}`
- Part selection creates a shallow copy of QuestionGroup with only selected parts

## Rendering
- LaTeX uses the `exam` document class
- Groups render with `\uplevel{stem...}` then individual `\question` for each part
- `sectionItem` interface (private to `exam` pkg): `renderTeX() string`
  - `*renderQuestion` — single question
  - `*renderGroup` — uplevel block + parts
- Group metadata (`\footnotesize`) shown in `\uplevel` block when `ShowMetadata=true`
- Parts never show metadata annotation (it's on the group's uplevel instead)
- `buildRenderQuestion` helper avoids duplication between standalone and group-part rendering

## Key files
- `question/question.go`: `Question`, `Parse`, `ParseFile`, `LoadBank`, `postProcess`
- `question/group.go`: `BankItem`, `Bank`, `QuestionGroup`, `ParseGroup`, `ParseGroupFile`
- `exam/exam.go`: `Exam`, `SectionItem`, `Section`, `Resolve`
- `exam/render.go`: `Render`, `renderQuestion`, `renderGroup`, `buildRenderQuestion`
- `exam/exam.tmpl`: LaTeX template
- `testdata/bank/`: sample questions and `processes-group-001.group.toml`
- `testdata/exams/exam.toml`: sample exam using new `SectionItem` format
