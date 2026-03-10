/*
Copyright © 2026 Tej Chajed <tchajed@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tchajed/question-bank/question"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate all questions in the bank and report errors",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		bank, loadErr := question.LoadBank(bankDir)

		// Count successfully loaded items
		questionCount := 0
		groupCount := 0
		for _, item := range bank {
			switch item.(type) {
			case *question.Question:
				questionCount++
			case *question.QuestionGroup:
				groupCount++
			}
		}

		// Collect parse errors
		var parseErrors []error
		if loadErr != nil {
			if joined, ok := loadErr.(interface{ Unwrap() []error }); ok {
				parseErrors = joined.Unwrap()
			} else {
				parseErrors = []error{loadErr}
			}
		}

		// Run deep validation on successfully loaded questions
		warnings := question.ValidateBank(bank)

		// Print parse errors
		if len(parseErrors) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), "Errors:")
			for _, e := range parseErrors {
				fmt.Fprintf(cmd.ErrOrStderr(), "  %s\n", e)
			}
		}

		// Print warnings
		if len(warnings) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), "Warnings:")
			for _, w := range warnings {
				fmt.Fprintf(cmd.ErrOrStderr(), "  %s\n", w)
			}
		}

		// Print summary
		fmt.Fprintf(cmd.OutOrStdout(), "Validated %d questions, %d groups", questionCount, groupCount)
		if len(parseErrors) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), ", %d errors", len(parseErrors))
		}
		if len(warnings) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), ", %d warnings", len(warnings))
		}
		if len(parseErrors) == 0 && len(warnings) == 0 {
			fmt.Fprint(cmd.OutOrStdout(), ", no errors")
		}
		fmt.Fprintln(cmd.OutOrStdout())

		if len(parseErrors) > 0 || len(warnings) > 0 {
			return fmt.Errorf("validation failed")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
