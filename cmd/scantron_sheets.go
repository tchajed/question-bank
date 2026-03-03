/*
Copyright © 2026 Tej Chajed <tchajed@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/tchajed/question-bank/exam"
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

		e, resolved, absBankDir, err := loadAndResolve(sheetsExamPath, bankDir)
		if err != nil {
			return err
		}

		absOutputDir, err := filepath.Abs(sheetsOutputDir)
		if err != nil {
			return err
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

		// Render each student's sheet in parallel.
		type job struct {
			rec    *scantron.StudentRecord
			graded *scantron.GradedRecord
		}

		jobs := make(chan job, len(data.Records))
		for _, rec := range data.Records {
			graded, err := scantron.Grade(rec, key)
			if err != nil {
				return fmt.Errorf("grading student %s: %w", rec.ID, err)
			}
			jobs <- job{rec, graded}
		}
		close(jobs)

		type result struct {
			pdfDst      string
			displayName string
			err         error
		}
		results := make(chan result, len(data.Records))

		var wg sync.WaitGroup
		numWorkers := runtime.NumCPU()
		for range numWorkers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := range jobs {
					rec, graded := j.rec, j.graded
					student := exam.StudentResponse{
						Name:      fmt.Sprintf("%s, %s", rec.LastName, rec.FirstName),
						ID:        rec.ID,
						Responses: rec.Responses,
						Earned:    graded.NumCorrect,
						Total:     graded.NumTotal,
					}

					latex, err := e.RenderStudentSheet(resolved, absBankDir, student)
					if err != nil {
						results <- result{err: fmt.Errorf("rendering sheet for %s: %w", rec.ID, err)}
						continue
					}

					pdfDst, err := compileStudentPDF(latex, rec.ID, absOutputDir)
					if err != nil {
						results <- result{err: err}
						continue
					}

					displayName := strings.ReplaceAll(student.Name, `\`, "")
					results <- result{pdfDst: pdfDst, displayName: displayName}
				}
			}()
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		for r := range results {
			if r.err != nil {
				return r.err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "wrote %s (%s)\n", r.pdfDst, r.displayName)
		}

		return nil
	},
}

// compileStudentPDF compiles LaTeX source into a PDF via latexmk in a
// temporary directory, copies the result to outputDir/<id>.pdf, and returns
// the destination path.
func compileStudentPDF(latex []byte, id, outputDir string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "qb-sheet-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	texFile := id + ".tex"
	if err := os.WriteFile(filepath.Join(tmpDir, texFile), latex, 0o644); err != nil {
		return "", fmt.Errorf("writing tex file: %w", err)
	}

	latexmk := exec.Command("latexmk", "-pdf", "-interaction=nonstopmode", texFile)
	latexmk.Dir = tmpDir
	out, err := latexmk.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("latexmk failed for %s: %w\n%s", id, err, out)
	}

	pdfDst := filepath.Join(outputDir, id+".pdf")
	if err := copyFile(filepath.Join(tmpDir, id+".pdf"), pdfDst); err != nil {
		return "", fmt.Errorf("copying PDF for %s: %w", id, err)
	}
	return pdfDst, nil
}

func init() {
	scantronCmd.AddCommand(scantronSheetsCmd)
	scantronSheetsCmd.Flags().StringVar(&sheetsExamPath, "exam", "", "canonical exam TOML file (required)")
	scantronSheetsCmd.Flags().StringVarP(&sheetsOutputDir, "output-dir", "o", "./sheets", "output directory for PDFs")
	_ = scantronSheetsCmd.MarkFlagRequired("exam")
}
