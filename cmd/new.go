/*
Copyright © 2026 Tej Chajed <tchajed@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tchajed/question-bank/question"
)

var trueFalse bool
var shortAnswer bool
var fillInTheBlank bool

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

		var data []byte
		if trueFalse {
			data, err = question.TrueFalseTemplate()
		} else if shortAnswer {
			data, err = question.ShortAnswerTemplate()
		} else if fillInTheBlank {
			data, err = question.FillInTheBlankTemplate()
		} else {
			data, err = question.MultipleChoiceTemplate()
		}
		if err != nil {
			return err
		}

		if err := os.WriteFile(path, data, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "created %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().BoolVar(&trueFalse, "true-false", false, "create a true/false question template")
	newCmd.Flags().BoolVar(&shortAnswer, "short-answer", false, "create a short-answer question template")
	newCmd.Flags().BoolVar(&fillInTheBlank, "fill-in-the-blank", false, "create a fill-in-the-blank question template")
}
