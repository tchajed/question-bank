/*
Copyright © 2026 Tej Chajed <tchajed@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tchajed/question-bank/exam"
	"github.com/tchajed/question-bank/question"
	"github.com/tchajed/question-bank/scantron"
)

var sheetsExamPath string
var sheetsOutputDir string

var scantronSheetsCmd = &cobra.Command{
	Use:   "sheets <reordered.csv>",
	Short: "Render per-student exam sheets with color-coded answers",
	Long: `Reads a reordered scantron CSV and renders a personalized PDF for each
student showing their answers color-coded: correct in green, incorrect in red.

The CSV should already be in canonical question order (output of "qb scantron reorder").`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		csvPath := args[0]

		absExamPath, err := filepath.Abs(sheetsExamPath)
		if err != nil {
			return err
		}

		absBankDir, err := filepath.Abs(bankDir)
		if err != nil {
			return err
		}

		absOutputDir, err := filepath.Abs(sheetsOutputDir)
		if err != nil {
			return err
		}

		// Load exam and bank.
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

		// Derive answer key for score computation.
		key := scantron.DeriveAnswerKey(resolved)

		// Create output directory.
		if err := os.MkdirAll(absOutputDir, 0o755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}

		// Render each student's sheet.
		for _, rec := range data.Records {
			graded, err := scantron.Grade(rec, key)
			if err != nil {
				return fmt.Errorf("grading student %s: %w", rec.ID, err)
			}

			student := exam.StudentResponse{
				Name:      fmt.Sprintf("%s, %s", rec.LastName, rec.FirstName),
				ID:        rec.ID,
				Responses: rec.Responses,
				Earned:    graded.NumCorrect,
				Total:     graded.NumTotal,
			}

			latex, err := e.RenderStudentSheet(resolved, absBankDir, student)
			if err != nil {
				return fmt.Errorf("rendering sheet for %s: %w", rec.ID, err)
			}

			// Compile with latexmk in a temporary directory.
			tmpDir, err := os.MkdirTemp("", "qb-sheet-*")
			if err != nil {
				return fmt.Errorf("creating temp dir: %w", err)
			}

			texFile := rec.ID + ".tex"
			texPath := filepath.Join(tmpDir, texFile)
			if err := os.WriteFile(texPath, latex, 0o644); err != nil {
				os.RemoveAll(tmpDir)
				return fmt.Errorf("writing tex file: %w", err)
			}

			latexmk := exec.Command("latexmk", "-pdf", "-interaction=nonstopmode", texFile)
			latexmk.Dir = tmpDir
			out, latexErr := latexmk.CombinedOutput()
			if latexErr != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "%s", out)
				os.RemoveAll(tmpDir)
				return fmt.Errorf("latexmk failed for %s: %w", rec.ID, latexErr)
			}

			pdfSrc := filepath.Join(tmpDir, rec.ID+".pdf")
			pdfDst := filepath.Join(absOutputDir, rec.ID+".pdf")
			if err := copyFile(pdfSrc, pdfDst); err != nil {
				os.RemoveAll(tmpDir)
				return fmt.Errorf("copying PDF for %s: %w", rec.ID, err)
			}

			os.RemoveAll(tmpDir)

			// Sanitize student name for display: replace LaTeX escapes
			displayName := strings.ReplaceAll(student.Name, `\`, "")
			fmt.Fprintf(cmd.OutOrStdout(), "wrote %s (%s)\n", pdfDst, displayName)
		}

		return nil
	},
}

func init() {
	scantronCmd.AddCommand(scantronSheetsCmd)
	scantronSheetsCmd.Flags().StringVar(&sheetsExamPath, "exam", "", "canonical exam TOML file (required)")
	scantronSheetsCmd.Flags().StringVarP(&sheetsOutputDir, "output-dir", "o", "./sheets", "output directory for PDFs")
	_ = scantronSheetsCmd.MarkFlagRequired("exam")
}
