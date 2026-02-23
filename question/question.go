// question implements the parsing and serialization of questions
package question

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Difficulty represents the difficulty level of a question.
type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

// QuestionType identifies the format of a question.
type QuestionType string

const (
	MultipleChoice QuestionType = "multiple-choice"
	TrueFalse      QuestionType = "true-false"
)

// Short returns an abbreviated form of the question type: "choice" or "tf".
func (t QuestionType) Short() string {
	if t == TrueFalse {
		return "tf"
	}
	if t == MultipleChoice {
		return "choice"
	}
	panic(fmt.Errorf("unhandled type: %s", t))
}

// Choice is one option in a multiple-choice question.
type Choice struct {
	Text    string `toml:"text"`
	Correct bool   `toml:"correct,omitempty"`
}

// Question represents a single exam/quiz question.
//
// Required fields: Topic and Stem.
type Question struct {
	// Id is derived from the file path (relative path without extension).
	Id string `toml:"-"`
	// Question prompt
	Stem    string       `toml:"stem"`
	Type    QuestionType `toml:"type"`
	Choices []Choice     `toml:"choices"`
	// Explanation of correct answer for solutions
	Explanation string `toml:"explanation"`
	// Answer for true-false questions
	AnswerTF *bool `toml:"answer_tf,omitempty"`

	// Topic helps categorize questions. Can be hierarchical, separated by '/'.
	Topic string `toml:"topic"`
	// Difficulty is easy/medium/hard
	Difficulty Difficulty `toml:"difficulty"`
	// Tags are keywords to categorize and find questions
	Tags []string `toml:"tags"`
	// (Optional) figure to include alongside question stem.
	Figure string `toml:"figure"`
	// Optional: treated as 1 if 0 or omitted
	Points int `toml:"points,omitempty"`
}

// validate checks that required fields are present and consistent.
func (q *Question) validate() error {
	if q.Topic == "" {
		return fmt.Errorf("question missing required field: topic")
	}
	if q.Stem == "" {
		return fmt.Errorf("question missing required field: stem")
	}
	if q.Type == TrueFalse && q.AnswerTF == nil {
		return fmt.Errorf("true-false question missing required field: answer")
	}
	return nil
}

// postProcess validates and normalizes a decoded Question. It is called after
// TOML decoding, and also after inheriting group metadata for group parts.
func postProcess(q *Question) error {
	if err := q.validate(); err != nil {
		return err
	}
	if q.Type == "" {
		if q.AnswerTF != nil {
			q.Type = TrueFalse
		} else {
			q.Type = MultipleChoice
		}
	}
	if q.Type == TrueFalse {
		q.Choices = []Choice{
			{Text: "True", Correct: *q.AnswerTF},
			{Text: "False", Correct: !*q.AnswerTF},
		}
	}
	if q.Points == 0 {
		q.Points = 1
	}
	return nil
}

// GetId returns the question's unique identifier.
func (q *Question) GetId() string { return q.Id }

// Parse parses a Question from TOML-encoded bytes.
func Parse(data []byte) (*Question, error) {
	var q Question
	dec := toml.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&q); err != nil {
		return nil, err
	}
	if err := postProcess(&q); err != nil {
		return nil, err
	}
	return &q, nil
}

// LoadBank reads all questions and question groups from a directory tree,
// returning a Bank indexed by ID. Files ending in .group.toml are parsed as
// QuestionGroups; other .toml files are parsed as Questions. Files that fail
// to parse are collected and returned as a combined error; successfully parsed
// items are still returned.
func LoadBank(dir string) (Bank, error) {
	bank := make(Bank)
	var errs []error
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			errs = append(errs, err)
			return nil
		}
		if d.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			errs = append(errs, err)
			return nil
		}
		if strings.HasSuffix(path, ".group.toml") {
			g, err := ParseGroupFile(dir, relPath)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", relPath, err))
				return nil
			}
			bank[g.Id] = g
			for _, part := range g.Parts {
				bank[part.Id] = part
			}
		} else if filepath.Ext(path) == ".toml" {
			q, err := ParseFile(dir, relPath)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", relPath, err))
				return nil
			}
			bank[q.Id] = q
		}
		return nil
	})
	if err != nil {
		return bank, err
	}
	return bank, errors.Join(errs...)
}

// ParseFile reads and parses a Question from a TOML file.
//
// baseDir is the root directory of the question bank, and relPath is the path
// to the file relative to baseDir. The question's ID is set to relPath without
// its file extension.
func ParseFile(baseDir, relPath string) (*Question, error) {
	data, err := os.ReadFile(filepath.Join(baseDir, relPath))
	if err != nil {
		return nil, err
	}
	q, err := Parse(data)
	if err != nil {
		return nil, err
	}
	ext := filepath.Ext(relPath)
	q.Id = strings.TrimSuffix(relPath, ext)
	return q, nil
}
