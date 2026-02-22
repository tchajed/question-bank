/*
Copyright © 2026 Tej Chajed <tchajed@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tchajed/question-bank/exam"
	"github.com/tchajed/question-bank/question"
)

var renderOutput string

var renderCmd = &cobra.Command{
	Use:   "render <exam.toml>",
	Short: "Render an exam to LaTeX",
	Long: `Render an exam TOML file to a LaTeX document.

Reads defaults.toml from the same directory as the exam file (if present) for
course-level settings such as course_code and cover_page.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		examPath := args[0]

		absExamPath, err := filepath.Abs(examPath)
		if err != nil {
			return err
		}
		examDir := filepath.Dir(absExamPath)

		absBankDir, err := filepath.Abs(bankDir)
		if err != nil {
			return err
		}

		e, err := exam.LoadWithDefaults(absExamPath)
		if err != nil {
			return fmt.Errorf("loading exam: %w", err)
		}

		bank, err := question.LoadBank(absBankDir)
		if err != nil {
			return fmt.Errorf("loading bank: %w", err)
		}

		resolved, err := e.Resolve(bank)
		if err != nil {
			return fmt.Errorf("resolving questions: %w", err)
		}

		latex, err := e.Render(resolved, absBankDir, examDir)
		if err != nil {
			return fmt.Errorf("rendering: %w", err)
		}

		outPath := renderOutput
		if outPath == "" {
			base := strings.TrimSuffix(filepath.Base(absExamPath), ".toml")
			outPath = filepath.Join(examDir, base+".tex")
		}

		if outPath == "-" {
			_, err = cmd.OutOrStdout().Write(latex)
			return err
		}

		if err := os.WriteFile(outPath, latex, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", outPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(renderCmd)
	renderCmd.Flags().StringVarP(&renderOutput, "output", "o", "", "output file (default: exam name with .tex extension; use - for stdout)")
}
