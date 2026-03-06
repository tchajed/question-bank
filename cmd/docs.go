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
Pass "question" or "exam" to print only that document.

Use --prompt to output a self-contained LLM prompt that can be fed to any
coding assistant to help convert existing exam content into question-bank
TOML files.`,
	Args:      cobra.MaximumNArgs(1),
	ValidArgs: []string{"question", "exam"},
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt, _ := cmd.Flags().GetBool("prompt")
		if prompt {
			fmt.Fprint(cmd.OutOrStdout(), docs.ImportPrompt())
			return nil
		}

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
	docsCmd.Flags().Bool("prompt", false, "output a self-contained LLM prompt for importing exam content into question-bank TOML files")
	rootCmd.AddCommand(docsCmd)
}
