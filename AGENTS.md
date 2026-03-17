# question-bank

A CLI app to create quizzes/exams in a structured format. Questions are stored in individual TOML files, and exams are written as TOML files that reference questions. Supports rendering exams to PDF and exporting to Canvas with a QTI zip.

## Commands

```bash
go test ./...
go mod tidy
```

Run the tests to validate your work.

## Architecture

This is a Go library (`github.com/tchajed/question-bank`) for managing exam/quiz questions stored as TOML files.

There is a CLI implemented using spf13/cobra with subcommands in cmd/.

## File format

The question schema is in `./question/question.go` and the exam schema is in `./exam/exam.go`. Use `docs/question-format.md` and `docs/exam-format.md` as your primary guides, and keep them up-to-date as the format changes.
