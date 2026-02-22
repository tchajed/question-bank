/*
Copyright © 2026 Tej Chajed <tchajed@gmail.com>
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tchajed/question-bank/exam"
	"github.com/tchajed/question-bank/question"
)

var renderOutput string
var renderTexOnly bool
var renderSolution bool
var renderMetadata bool

var renderCmd = &cobra.Command{
	Use:   "render <exam.toml>",
	Short: "Render an exam to PDF (or LaTeX with --tex)",
	Long: `Render an exam TOML file to a PDF by compiling with latexmk.

Reads defaults.toml from the same directory as the exam file (if present) for
course-level settings such as course_code and cover_page.

The generated PDF is placed next to the exam TOML file. Use --tex to generate
only the LaTeX source file instead (useful for debugging).`,
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

		latex, err := e.Render(resolved, absBankDir, exam.RenderOptions{
			PrintAnswers: renderSolution,
			ShowMetadata: renderMetadata,
		})
		if err != nil {
			return fmt.Errorf("rendering: %w", err)
		}

		base := strings.TrimSuffix(filepath.Base(absExamPath), ".toml")

		if renderTexOnly {
			outPath := renderOutput
			if outPath == "" {
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
		}

		// Compile with latexmk in a temporary directory.
		tmpDir, err := os.MkdirTemp("", "question-bank-*")
		if err != nil {
			return fmt.Errorf("creating temp dir: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		texFile := base + ".tex"
		texPath := filepath.Join(tmpDir, texFile)
		if err := os.WriteFile(texPath, latex, 0o644); err != nil {
			return fmt.Errorf("writing tex file: %w", err)
		}

		latexmk := exec.Command("latexmk", "-pdf", "-interaction=nonstopmode", texFile)
		latexmk.Dir = tmpDir
		out, err := latexmk.CombinedOutput()
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "%s", out)
			return fmt.Errorf("latexmk failed: %w", err)
		}

		pdfSrc := filepath.Join(tmpDir, base+".pdf")
		pdfDst := renderOutput
		if pdfDst == "" {
			pdfDst = filepath.Join(examDir, base+".pdf")
		}

		if err := copyFile(pdfSrc, pdfDst); err != nil {
			return fmt.Errorf("copying PDF: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", pdfDst)
		return nil
	},
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func init() {
	rootCmd.AddCommand(renderCmd)
	renderCmd.Flags().StringVarP(&renderOutput, "output", "o", "", "output file path (default: exam name with .pdf or .tex extension next to the TOML file; use - for stdout with --tex)")
	renderCmd.Flags().BoolVar(&renderTexOnly, "tex", false, "generate .tex file only (for debugging)")
	renderCmd.Flags().BoolVar(&renderSolution, "solution", false, "include answers and solutions in the output")
	renderCmd.Flags().BoolVar(&renderMetadata, "metadata", false, "render question metadata (ID, topic, difficulty) before each question")
}
