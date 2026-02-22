/*
Copyright © 2026 Tej Chajed <tchajed@gmail.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var bankDir string

var rootCmd = &cobra.Command{
	Use:   "question-bank",
	Short: "Manage a question bank",
	Long:  `question-bank is a CLI tool for managing exam and quiz questions stored as TOML files.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&bankDir, "bank", "b", ".", "question bank directory")
}
