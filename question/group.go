package question

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// BankItem is implemented by *Question and *QuestionGroup.
// Use a type switch to access type-specific fields.
type BankItem interface {
	GetId() string
}

// Bank maps item IDs to their contents.
type Bank map[string]BankItem

// QuestionGroup is a multi-part question with shared instructions and metadata.
// Sub-questions (Parts) each get an ID of the form "group-id/N" (1-indexed).
type QuestionGroup struct {
	// Id is derived from the file path (without .group.toml extensions).
	Id string `toml:"-"`
	// Stem is the shared instructions / scenario for all parts.
	Stem   string `toml:"stem"`
	Figure string `toml:"figure,omitempty"`

	// Metadata shared across parts (parts inherit these if not set explicitly).
	Topic      string     `toml:"topic"`
	Difficulty Difficulty `toml:"difficulty,omitempty"`
	Tags       []string   `toml:"tags,omitempty"`

	// Parts are the individual sub-questions.
	Parts []*Question `toml:"parts"`
}

// GetId returns the group's unique identifier.
func (g *QuestionGroup) GetId() string { return g.Id }

// ParseGroup parses a QuestionGroup from TOML-encoded bytes.
// Each part inherits Topic, Difficulty, and Tags from the group if not set.
func ParseGroup(data []byte) (*QuestionGroup, error) {
	var g QuestionGroup
	dec := toml.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&g); err != nil {
		return nil, err
	}
	if g.Topic == "" {
		return nil, fmt.Errorf("question group missing required field: topic")
	}
	if g.Stem == "" {
		return nil, fmt.Errorf("question group missing required field: stem")
	}
	if len(g.Parts) == 0 {
		return nil, fmt.Errorf("question group must have at least one part")
	}
	for i, part := range g.Parts {
		if part.Topic == "" {
			part.Topic = g.Topic
		}
		if part.Difficulty == "" {
			part.Difficulty = g.Difficulty
		}
		if len(part.Tags) == 0 {
			part.Tags = g.Tags
		}
		if err := postProcess(part); err != nil {
			return nil, fmt.Errorf("part %d: %w", i+1, err)
		}
	}
	return &g, nil
}

// ParseGroupFile reads and parses a QuestionGroup from a .group.toml file.
//
// baseDir is the root directory of the question bank and relPath is the path
// to the file relative to baseDir. The group's ID is set to relPath with
// the .group.toml suffix stripped, and each part's ID is set to "group-id/N"
// (1-indexed).
func ParseGroupFile(baseDir, relPath string) (*QuestionGroup, error) {
	data, err := os.ReadFile(filepath.Join(baseDir, relPath))
	if err != nil {
		return nil, err
	}
	g, err := ParseGroup(data)
	if err != nil {
		return nil, err
	}
	id := strings.TrimSuffix(relPath, ".group.toml")
	g.Id = id
	for i, part := range g.Parts {
		part.Id = fmt.Sprintf("%s/%d", id, i+1)
	}
	return g, nil
}
