/*
Copyright © 2026 Tej Chajed <tchajed@gmail.com>
*/
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

// questionTemplate is used only for generating a commented template file.
type questionTemplate struct {
	Stem        string           `toml:"stem,multiline" comment:"Required. The question prompt."`
	Topic       string           `toml:"topic" comment:"Used to organize questions."`
	Answer      *bool            `toml:"answer,omitempty" comment:"Correct answer (true/false)"`
	Explanation string           `toml:"explanation,multiline" comment:"Explanation of the answer for solutions."`
	Choices     []choiceTemplate `toml:"choices,omitempty" comment:"Answer choices"`

	Type       string   `toml:"type" comment:"Question type: 'multiple-choice' (default) or 'true-false'"`
	Difficulty string   `toml:"difficulty"`
	Tags       []string `toml:"tags" comment:"Keywords to categorize and find questions"`
	Figure     string   `toml:"figure,omitempty,commented" comment:"Optional figure path to include alongside the question stem"`
	Points     int      `toml:"points,omitempty,commented" comment:"Point value; treated as 1 if omitted"`
}

type choiceTemplate struct {
	Text    string `toml:"text"`
	Correct bool   `toml:"correct,omitempty"`
}

var trueFalse bool

var newCmd = &cobra.Command{
	Use:   "new <file>",
	Short: "Create a new question template file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		absBank, err := filepath.Abs(bankDir)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(absBank, absPath)
		if err != nil || !filepath.IsLocal(rel) {
			return fmt.Errorf("%s is not within the question bank (%s)", path, bankDir)
		}

		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists", path)
		}

		tmpl := questionTemplate{
			Difficulty:  "medium",
			Tags:        []string{},
			Stem:        "\n",
			Explanation: "\n",
			Points:      1,
		}
		if trueFalse {
			tmpl.Type = "true-false"
			f := false
			tmpl.Answer = &f
		} else {
			tmpl.Type = "multiple-choice"
			tmpl.Choices = []choiceTemplate{
				{Text: "", Correct: true},
				{Text: ""},
				{Text: ""},
			}
		}

		var buf bytes.Buffer
		enc := toml.NewEncoder(&buf)
		enc.SetArraysMultiline(true)
		if err := enc.Encode(tmpl); err != nil {
			return err
		}

		if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "created %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().BoolVar(&trueFalse, "true-false", false, "create a true/false question template")
}
