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
	"github.com/tchajed/question-bank/scantron"
)

var analysisExam string
var analysisFormat string
var analysisOutput string

var scantronAnalysisCmd = &cobra.Command{
	Use:   "analysis <reordered.csv>",
	Short: "Compute per-question item analysis by score quintile",
	Long: `Compute per-question statistics grouped by score quintile from a
reordered scantron CSV. Students are sorted by total score and divided
into 5 quintile groups. For each question, the output includes overall
percent correct and per-quintile percent correct.

The CSV output contains: Question, ID, OverallPctCorrect, and Q1-Q5 percent
correct. The JSON output contains full statistics including response
distributions per quintile.`,
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

		absExam, err := filepath.Abs(analysisExam)
		if err != nil {
			return err
		}
		e, err := exam.LoadWithDefaults(absExam)
		if err != nil {
			return fmt.Errorf("loading exam: %w", err)
		}
		resolved, err := e.Resolve(bank)
		if err != nil {
			return fmt.Errorf("resolving exam: %w", err)
		}

		// Parse CSV.
		csvFile, err := os.Open(csvPath)
		if err != nil {
			return fmt.Errorf("opening CSV: %w", err)
		}
		defer csvFile.Close()

		data, err := scantron.ParseCSV(csvFile)
		if err != nil {
			return fmt.Errorf("parsing CSV: %w", err)
		}

		// Derive answer key and metadata from the resolved exam.
		key := scantron.DeriveAnswerKey(resolved)
		questionIDs := scantron.QuestionIDs(resolved)
		numChoices := scantron.NumChoices(resolved)

		// Run analysis.
		stats, err := scantron.ItemAnalysis(data.Records, key, questionIDs, numChoices)
		if err != nil {
			return fmt.Errorf("running item analysis: %w", err)
		}

		// Write output.
		out := cmd.OutOrStdout()
		if analysisOutput != "" {
			f, err := os.Create(analysisOutput)
			if err != nil {
				return err
			}
			defer f.Close()
			out = f
		}

		switch analysisFormat {
		case "csv":
			return scantron.WriteAnalysisCSV(out, stats)
		case "json":
			return scantron.WriteAnalysisJSON(out, stats)
		default:
			return fmt.Errorf("unknown format %q, expected csv or json", analysisFormat)
		}
	},
}

func init() {
	scantronCmd.AddCommand(scantronAnalysisCmd)
	scantronAnalysisCmd.Flags().StringVar(&analysisExam, "exam", "", "canonical exam TOML file (required)")
	scantronAnalysisCmd.Flags().StringVar(&analysisFormat, "format", "csv", "output format: csv or json")
	scantronAnalysisCmd.Flags().StringVarP(&analysisOutput, "output", "o", "", "output file path (default: stdout)")
	_ = scantronAnalysisCmd.MarkFlagRequired("exam")
}
