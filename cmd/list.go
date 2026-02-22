/*
Copyright © 2026 Tej Chajed <tchajed@gmail.com>
*/
package cmd

import (
	"fmt"
	"maps"
	"slices"

	"github.com/spf13/cobra"
	"github.com/tchajed/question-bank/question"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all questions in the bank",
	RunE: func(cmd *cobra.Command, args []string) error {
		bank, err := question.LoadBank(bankDir)
		if err != nil {
			return err
		}
		ids := slices.Sorted(maps.Keys(bank))
		for _, id := range ids {
			q := bank[id]
			fmt.Fprintf(cmd.OutOrStdout(), "%-30s  %-6s  %-10s  %s\n", q.Id, q.Type.Short(), q.Difficulty, q.Topic)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
