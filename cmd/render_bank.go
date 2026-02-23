/*
Copyright © 2026 Tej Chajed <tchajed@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tchajed/question-bank/exam"
	"github.com/tchajed/question-bank/question"
)

var renderBankOutput string
var renderBankTexOnly bool

const renderBankSolution = true
const renderBankMetadata = true

var renderBankCmd = &cobra.Command{
	Use:   "render-bank",
	Short: "Render the entire question bank to PDF (or LaTeX with --tex)",
	Long: `Render all questions in the bank as a single exam PDF.

All top-level questions and groups are included in a single section, sorted by
ID. Use --tex to generate only the LaTeX source file instead.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		absBankDir, err := filepath.Abs(bankDir)
		if err != nil {
			return err
		}

		bank, err := question.LoadBank(absBankDir)
		if err != nil {
			return fmt.Errorf("loading bank: %w", err)
		}

		e := exam.BankAsExam(bank)

		resolved, err := e.Resolve(bank)
		if err != nil {
			return fmt.Errorf("resolving questions: %w", err)
		}

		latex, err := e.Render(resolved, absBankDir, exam.RenderOptions{
			PrintAnswers: renderBankSolution,
			ShowMetadata: renderBankMetadata,
		})
		if err != nil {
			return fmt.Errorf("rendering: %w", err)
		}

		if renderBankTexOnly {
			outPath := renderBankOutput
			if outPath == "" {
				outPath = filepath.Join(absBankDir, "bank.tex")
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
		}

		// Compile with latexmk in a temporary directory.
		tmpDir, err := os.MkdirTemp("", "question-bank-*")
		if err != nil {
			return fmt.Errorf("creating temp dir: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		texPath := filepath.Join(tmpDir, "bank.tex")
		if err := os.WriteFile(texPath, latex, 0o644); err != nil {
			return fmt.Errorf("writing tex file: %w", err)
		}

		latexmk := exec.Command("latexmk", "-pdf", "-interaction=nonstopmode", "bank.tex")
		latexmk.Dir = tmpDir
		out, err := latexmk.CombinedOutput()
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "%s", out)
			return fmt.Errorf("latexmk failed: %w", err)
		}

		pdfDst := renderBankOutput
		if pdfDst == "" {
			pdfDst = filepath.Join(absBankDir, "bank.pdf")
		}
		if err := copyFile(filepath.Join(tmpDir, "bank.pdf"), pdfDst); err != nil {
			return fmt.Errorf("copying PDF: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", pdfDst)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(renderBankCmd)
	renderBankCmd.Flags().StringVarP(&renderBankOutput, "output", "o", "", "output file path (default: bank.pdf or bank.tex in bank dir; use - for stdout with --tex)")
	renderBankCmd.Flags().BoolVar(&renderBankTexOnly, "tex", false, "generate .tex file only (for debugging)")
}
