/*
Copyright © 2026 Tej Chajed <tchajed@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tchajed/question-bank/docs"
)

var docsCmd = &cobra.Command{
	Use:   "docs [question|exam]",
	Short: "Print format reference documentation",
	Long: `Print format reference documentation for the question bank file formats.

Without arguments, prints both the question format and exam format references.
Pass "question" or "exam" to print only that document.`,
	Args:      cobra.MaximumNArgs(1),
	ValidArgs: []string{"question", "exam"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			fmt.Fprint(cmd.OutOrStdout(), docs.QuestionFormat)
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprint(cmd.OutOrStdout(), docs.ExamFormat)
			return nil
		}
		switch args[0] {
		case "question":
			fmt.Fprint(cmd.OutOrStdout(), docs.QuestionFormat)
		case "exam":
			fmt.Fprint(cmd.OutOrStdout(), docs.ExamFormat)
		default:
			return fmt.Errorf("unknown doc %q: use \"question\" or \"exam\"", args[0])
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
}
