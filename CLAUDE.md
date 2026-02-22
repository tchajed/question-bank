# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go test ./...              # Run all tests
go mod tidy                # Clean up dependencies
```

Run the tests to validate your work.

## Architecture

This is a Go library (`github.com/tchajed/question-bank`) for managing exam/quiz questions stored as TOML files.

The question schema is in `./question/question.go` and the exam schema is in `./exam/exam.go`.
