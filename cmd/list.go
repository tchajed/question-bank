/*
Copyright © 2026 Tej Chajed <tchajed@gmail.com>
*/
package cmd

import (
	"fmt"
	"maps"
	"slices"

	"github.com/spf13/cobra"
	"github.com/tchajed/question-bank/exam"
	"github.com/tchajed/question-bank/question"
)

func listItem(item question.BankItem) string {
	switch q := item.(type) {
	case *question.Question:
		return fmt.Sprintf("%-30s  %-6s  %-10s  %s", q.Id, q.Type.Short(), q.Difficulty, q.Topic)
	case *question.QuestionGroup:
		return fmt.Sprintf("%-30s  %-6s  %-10s  %s", q.Id, "group", q.Difficulty, q.Topic)
	}
	panic("unhandled BankItem type")
}

var listCmd = &cobra.Command{
	Use:   "list [exam.toml]",
	Short: "List all questions in the bank, or the questions in an exam",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			e, err := exam.LoadWithDefaults(args[0])
			if err != nil {
				return fmt.Errorf("loading exam: %w", err)
			}
			bank, err := question.LoadBank(bankDir)
			if err != nil {
				return err
			}
			resolved, err := e.Resolve(bank)
			if err != nil {
				return err
			}
			n := 1
			for _, sec := range resolved.Sections {
				if sec.Name != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "\n%s\n", sec.Name)
				}
				for _, item := range sec.Items {
					fmt.Fprintf(cmd.OutOrStdout(), "%3d. %s\n", n, listItem(item))
					n++
				}
			}
			return nil
		}

		bank, err := question.LoadBank(bankDir)
		if err != nil {
			return err
		}
		ids := slices.Sorted(maps.Keys(bank))
		for _, id := range ids {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", listItem(bank[id]))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
