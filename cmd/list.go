/*
Copyright © 2026 Tej Chajed <tchajed@gmail.com>
*/
package cmd

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tchajed/question-bank/question"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all questions in the bank",
	RunE: func(cmd *cobra.Command, args []string) error {
		var errs []error
		filepath.WalkDir(bankDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				errs = append(errs, err)
				return nil
			}
			if d.IsDir() || filepath.Ext(path) != ".toml" {
				return nil
			}
			relPath, err := filepath.Rel(bankDir, path)
			if err != nil {
				errs = append(errs, err)
				return nil
			}
			q, err := question.ParseFile(bankDir, relPath)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", relPath, err))
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%-30s  %-10s  %s\n", q.Id, q.Difficulty, q.Topic)
			return nil
		})
		for _, e := range errs {
			fmt.Fprintln(cmd.ErrOrStderr(), e)
		}
		if len(errs) > 0 {
			return fmt.Errorf("%d validation error(s)", len(errs))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
