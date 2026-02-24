/*
Copyright © 2026 Tej Chajed <tchajed@gmail.com>
*/
package cmd

import (
	"encoding/csv"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tchajed/question-bank/exam"
	"github.com/tchajed/question-bank/question"
)

var answerKeyCmd = &cobra.Command{
	Use:   "answer-key <exam.toml>",
	Short: "Output an answer key CSV for an exam",
	Long:  `Output a CSV with question number and answer choice (A, B, C, ...) for each question in the exam.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		examPath := args[0]

		absExamPath, err := filepath.Abs(examPath)
		if err != nil {
			return err
		}

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

		w := csv.NewWriter(cmd.OutOrStdout())
		if err := w.Write([]string{"question", "answer"}); err != nil {
			return err
		}
		questionNum := 1
		for _, sec := range resolved.Sections {
			for _, item := range sec.Items {
				var questions []*question.Question
				switch q := item.(type) {
				case *question.Question:
					questions = []*question.Question{q}
				case *question.QuestionGroup:
					questions = q.Parts
				}
				for _, q := range questions {
					answer, err := answerLetter(q)
					if err != nil {
						return fmt.Errorf("question %s: %w", q.Id, err)
					}
					if err := w.Write([]string{strconv.Itoa(questionNum), answer}); err != nil {
						return err
					}
					questionNum++
				}
			}
		}
		w.Flush()
		return w.Error()
	},
}

// answerLetter returns the answer letter (A, B, C, ...) for a question, or the
// answer string directly for short-answer questions.
func answerLetter(q *question.Question) (string, error) {
	if q.Type == question.ShortAnswer {
		return q.Answer, nil
	}
	for i, c := range q.Choices {
		if c.Correct {
			return string(rune('A' + i)), nil
		}
	}
	return "", fmt.Errorf("no correct answer found")
}

func init() {
	rootCmd.AddCommand(answerKeyCmd)
}
