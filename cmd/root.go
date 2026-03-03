/*
Copyright © 2026 Tej Chajed <tchajed@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tchajed/question-bank/exam"
	"github.com/tchajed/question-bank/question"
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

// loadAndResolve loads an exam and question bank, resolves all question
// references, and returns the results along with the absolute bank directory
// path (needed for figure paths during rendering).
func loadAndResolve(examPath, bankDirPath string) (e *exam.Exam, resolved *exam.ResolvedExam, absBankDir string, err error) {
	absExamPath, err := filepath.Abs(examPath)
	if err != nil {
		return nil, nil, "", err
	}
	absBankDir, err = filepath.Abs(bankDirPath)
	if err != nil {
		return nil, nil, "", err
	}
	e, err = exam.LoadWithDefaults(absExamPath)
	if err != nil {
		return nil, nil, "", fmt.Errorf("loading exam: %w", err)
	}
	bank, err := question.LoadBank(absBankDir)
	if err != nil {
		return nil, nil, "", fmt.Errorf("loading bank: %w", err)
	}
	resolved, err = e.Resolve(bank)
	if err != nil {
		return nil, nil, "", fmt.Errorf("resolving questions: %w", err)
	}
	return e, resolved, absBankDir, nil
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&bankDir, "bank", "b", ".", "question bank directory")
}
