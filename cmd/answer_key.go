/*
Copyright © 2026 Tej Chajed <tchajed@gmail.com>
*/
package cmd

import (
	"encoding/csv"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tchajed/question-bank/exam"
	"github.com/tchajed/question-bank/question"
)

var numericAnswerKey bool
var rowAnswerKey bool

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

		// Collect all answers
		var numbers []string
		var answers []string
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
					var answer string
					var err error
					if numericAnswerKey {
						answer, err = answerNumber(q)
					} else {
						answer, err = answerLetter(q)
					}
					if err != nil {
						return fmt.Errorf("question %s: %w", q.Id, err)
					}
					numStr := strconv.Itoa(questionNum)
					if numericAnswerKey {
						numStr = "_" + numStr
					}
					numbers = append(numbers, numStr)
					answers = append(answers, answer)
					questionNum++
				}
			}
		}

		w := csv.NewWriter(cmd.OutOrStdout())
		if rowAnswerKey {
			if err := w.Write(numbers); err != nil {
				return err
			}
			if err := w.Write(answers); err != nil {
				return err
			}
		} else {
			if err := w.Write([]string{"question", "answer"}); err != nil {
				return err
			}
			for i := range numbers {
				if err := w.Write([]string{numbers[i], answers[i]}); err != nil {
					return err
				}
			}
		}
		w.Flush()
		return w.Error()
	},
}

// blanksAnswer returns the first accepted answer for each blank, joined by "; "
// in sorted order by blank name.
func blanksAnswer(q *question.Question) string {
	names := make([]string, 0, len(q.Blanks))
	for name := range q.Blanks {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, q.Blanks[name].Answers[0])
	}
	return strings.Join(parts, "; ")
}

// answerLetter returns the answer letter (A, B, C, ...) for a question, or the
// answer string directly for short-answer and fill-in-the-blank questions.
func answerLetter(q *question.Question) (string, error) {
	if q.Type == question.ShortAnswer {
		return q.Answer, nil
	}
	if q.Type == question.FillInTheBlank {
		return blanksAnswer(q), nil
	}
	idx := q.CorrectChoiceIndex()
	if idx == 0 {
		return "", fmt.Errorf("no correct answer found")
	}
	return string(rune('A' + idx - 1)), nil
}

// answerNumber returns the 1-based answer number for a question, or the answer
// string directly for short-answer and fill-in-the-blank questions.
func answerNumber(q *question.Question) (string, error) {
	if q.Type == question.ShortAnswer {
		return q.Answer, nil
	}
	if q.Type == question.FillInTheBlank {
		return blanksAnswer(q), nil
	}
	idx := q.CorrectChoiceIndex()
	if idx == 0 {
		return "", fmt.Errorf("no correct answer found")
	}
	return strconv.Itoa(idx), nil
}

func init() {
	answerKeyCmd.Flags().BoolVar(&numericAnswerKey, "numeric", false, "output numeric answers (1, 2, 3, ...) instead of letters (A, B, C, ...)")
	answerKeyCmd.Flags().BoolVar(&rowAnswerKey, "row", false, "output as two rows (question numbers then answers) instead of two columns")
	rootCmd.AddCommand(answerKeyCmd)
}
