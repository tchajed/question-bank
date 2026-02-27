package cmd

import (
	"fmt"
	"html"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tchajed/question-bank/exam"
	"github.com/tchajed/question-bank/qti"
	"github.com/tchajed/question-bank/question"
)

var canvasOutput string

var canvasCmd = &cobra.Command{
	Use:   "canvas <exam.toml> [exam2.toml ...]",
	Short: "Export one or more exams to a Canvas QTI zip file",
	Long: `Export one or more exam TOML files as a Canvas QTI zip for uploading to Canvas.
Each exam becomes a separate assessment in the zip.

Reads defaults.toml from the same directory as each exam file (if present) for
course-level settings such as course_code.

With a single exam file, the output defaults to the exam name with a .zip
extension next to the TOML file. With multiple exam files all in the same
directory foo/, the output defaults to foo.zip next to that directory. If the
exam files are in different directories, --output is required.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		absBankDir, err := filepath.Abs(bankDir)
		if err != nil {
			return err
		}

		bank, err := question.LoadBank(absBankDir)
		if err != nil {
			return fmt.Errorf("loading bank: %w", err)
		}

		// Resolve absolute paths and check all exam directories.
		absPaths := make([]string, len(args))
		for i, arg := range args {
			p, err := filepath.Abs(arg)
			if err != nil {
				return err
			}
			absPaths[i] = p
		}

		outPath := canvasOutput
		if outPath == "" {
			if len(absPaths) == 1 {
				base := strings.TrimSuffix(filepath.Base(absPaths[0]), ".toml")
				outPath = filepath.Join(filepath.Dir(absPaths[0]), base+".zip")
			} else {
				// All exam files must share the same directory.
				dir := filepath.Dir(absPaths[0])
				for _, p := range absPaths[1:] {
					if filepath.Dir(p) != dir {
						return fmt.Errorf("exam files are in different directories; use --output to specify an output path")
					}
				}
				outPath = dir + ".zip"
			}
		}

		var quizzes []*qti.NewQuiz
		for _, absExamPath := range absPaths {
			e, err := exam.LoadWithDefaults(absExamPath)
			if err != nil {
				return fmt.Errorf("loading %s: %w", absExamPath, err)
			}

			resolved, err := e.Resolve(bank)
			if err != nil {
				return fmt.Errorf("resolving questions in %s: %w", absExamPath, err)
			}

			quizzes = append(quizzes, examToQuiz(e, resolved))
		}

		if err := qti.WriteZip(outPath, quizzes...); err != nil {
			return fmt.Errorf("writing QTI zip: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", outPath)
		return nil
	},
}

func examToQuiz(e *exam.Exam, resolved *exam.ResolvedExam) *qti.NewQuiz {
	var items []qti.NewItem
	var totalPoints float64

	for _, sec := range resolved.Sections {
		for _, bankItem := range sec.Items {
			switch v := bankItem.(type) {
			case *question.Question:
				item := questionToItem(v, "")
				items = append(items, item)
				totalPoints += item.Points
			case *question.QuestionGroup:
				for _, part := range v.Parts {
					item := questionToItem(part, v.Stem)
					items = append(items, item)
					totalPoints += item.Points
				}
			}
		}
	}

	title := e.Title
	if e.CourseCode != "" && e.Title != "" {
		title = e.CourseCode + " " + e.Title
	} else if e.CourseCode != "" {
		title = e.CourseCode
	}

	return &qti.NewQuiz{
		Title:          title,
		PointsPossible: totalPoints,
		Items:          items,
	}
}

// questionToItem converts a Question to a qti.NewItem.
// If groupStem is non-empty, it is prepended to the question stem to provide shared context.
func questionToItem(q *question.Question, groupStem string) qti.NewItem {
	stem := q.Stem
	if groupStem != "" {
		stem = groupStem + "\n\n" + stem
	}
	text := "<p>" + html.EscapeString(stem) + "</p>"

	var qtype qti.ItemType
	switch q.Type {
	case question.TrueFalse:
		qtype = qti.TrueFalseQuestion
	case question.MultipleChoice:
		correctCount := 0
		for _, c := range q.Choices {
			if c.Correct {
				correctCount++
			}
		}
		if correctCount > 1 {
			qtype = qti.MultipleAnswersQuestion
		} else {
			qtype = qti.MultipleChoiceQuestion
		}
	case question.ShortAnswer:
		qtype = qti.ShortAnswerQuestion
	}

	choices := make([]qti.NewChoice, len(q.Choices))
	for i, c := range q.Choices {
		choices[i] = qti.NewChoice{
			Text:    c.Text,
			Correct: c.Correct,
		}
	}

	var generalFeedback string
	if q.Explanation != "" {
		generalFeedback = "<p>" + html.EscapeString(q.Explanation) + "</p>"
	}

	return qti.NewItem{
		Title:           q.Id,
		Text:            text,
		Type:            qtype,
		Points:          float64(q.Points),
		Choices:         choices,
		GeneralFeedback: generalFeedback,
		Answer:          q.Answer,
	}
}

func init() {
	rootCmd.AddCommand(canvasCmd)
	canvasCmd.Flags().StringVarP(&canvasOutput, "output", "o", "", "output zip file path")
}
