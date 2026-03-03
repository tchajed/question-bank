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
	"github.com/tchajed/question-bank/scantron"
)

var reorderCanonical string
var reorderVersions []string
var reorderOutput string

var scantronReorderCmd = &cobra.Command{
	Use:   "reorder <results.csv>",
	Short: "Reorder scantron answers to canonical question order",
	Long: `Reorder answer columns in a scantron CSV so that all exam versions
map to a canonical question order.

The canonical exam defines the target question order. Each --version flag
maps a SpecialCodes value to an exam file with a different question order.
Students whose SpecialCodes don't match any --version flag are assumed to
already be in canonical order.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		csvPath := args[0]

		absBankDir, err := filepath.Abs(bankDir)
		if err != nil {
			return err
		}

		bank, err := question.LoadBank(absBankDir)
		if err != nil {
			return fmt.Errorf("loading bank: %w", err)
		}

		// Load canonical exam
		absCanonical, err := filepath.Abs(reorderCanonical)
		if err != nil {
			return err
		}
		canonExam, err := exam.LoadWithDefaults(absCanonical)
		if err != nil {
			return fmt.Errorf("loading canonical exam: %w", err)
		}
		canonResolved, err := canonExam.Resolve(bank)
		if err != nil {
			return fmt.Errorf("resolving canonical exam: %w", err)
		}

		// Build version map
		versions := make(scantron.VersionMap)
		for _, v := range reorderVersions {
			parts := strings.SplitN(v, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid --version format %q, expected CODE=path.toml", v)
			}
			code, examPath := parts[0], parts[1]

			absExamPath, err := filepath.Abs(examPath)
			if err != nil {
				return err
			}
			versionExam, err := exam.LoadWithDefaults(absExamPath)
			if err != nil {
				return fmt.Errorf("loading version %s exam: %w", code, err)
			}
			versionResolved, err := versionExam.Resolve(bank)
			if err != nil {
				return fmt.Errorf("resolving version %s exam: %w", code, err)
			}

			perm, err := scantron.DerivePermutation(canonResolved, versionResolved)
			if err != nil {
				return fmt.Errorf("deriving permutation for version %s: %w", code, err)
			}
			versions[code] = perm
		}

		// Parse CSV
		csvFile, err := os.Open(csvPath)
		if err != nil {
			return fmt.Errorf("opening CSV: %w", err)
		}
		defer csvFile.Close()

		data, err := scantron.ParseCSV(csvFile)
		if err != nil {
			return fmt.Errorf("parsing CSV: %w", err)
		}

		// Reorder
		reordered, err := scantron.ReorderAll(data, versions)
		if err != nil {
			return fmt.Errorf("reordering: %w", err)
		}

		// Write output
		out := cmd.OutOrStdout()
		if reorderOutput != "" {
			f, err := os.Create(reorderOutput)
			if err != nil {
				return err
			}
			defer f.Close()
			out = f
		}

		return scantron.WriteCSV(out, reordered)
	},
}

func init() {
	scantronCmd.AddCommand(scantronReorderCmd)
	scantronReorderCmd.Flags().StringVar(&reorderCanonical, "canonical", "", "exam TOML file for the canonical question order (required)")
	scantronReorderCmd.Flags().StringArrayVar(&reorderVersions, "version", nil, "map SpecialCodes to exam file: CODE=path.toml (repeatable)")
	scantronReorderCmd.Flags().StringVarP(&reorderOutput, "output", "o", "", "output CSV path (default: stdout)")
	_ = scantronReorderCmd.MarkFlagRequired("canonical")
}
