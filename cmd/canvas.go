package cmd

import (
	"fmt"
	"html"
	"path/filepath"
	"regexp"
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

	// questionNum tracks the current Canvas question number (text-only
	// items don't count as questions in Canvas numbering).
	questionNum := 1
	for i, sec := range resolved.Sections {
		if sec.Name != "" {
			items = append(items, qti.NewItem{
				Title: fmt.Sprintf("section-%d", i+1),
				Text:  "<h2>" + html.EscapeString(sec.Name) + "</h2>",
				Type:  qti.TextNoQuestion,
			})
		}
		for _, bankItem := range sec.Items {
			switch v := bankItem.(type) {
			case *question.Question:
				item := questionToItem(v)
				items = append(items, item)
				totalPoints += item.Points
				questionNum++
			case *question.QuestionGroup:
				firstNum := questionNum
				lastNum := questionNum + len(v.Parts) - 1
				groupStem := replaceGroupRefs(v.Stem, firstNum, lastNum)
				items = append(items, qti.NewItem{
					Title: v.Id,
					Text:  exam.MarkdownToHTML(groupStem),
					Type:  qti.TextNoQuestion,
				})
				for _, part := range v.Parts {
					item := questionToItem(part)
					items = append(items, item)
					totalPoints += item.Points
					questionNum++
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

var latexRefRe = regexp.MustCompile(`\\ref\{[^}]*\}`)

// replaceGroupRefs replaces \ref{GROUP:first} and \ref{GROUP:last} in the
// group stem with actual question numbers for the Canvas export.
func replaceGroupRefs(stem string, first, last int) string {
	stem = strings.ReplaceAll(stem, `\ref{GROUP:first}`, fmt.Sprintf("%d", first))
	stem = strings.ReplaceAll(stem, `\ref{GROUP:last}`, fmt.Sprintf("%d", last))
	// Strip any remaining \ref{...} that won't resolve in HTML.
	stem = latexRefRe.ReplaceAllStringFunc(stem, func(m string) string {
		return m[len(`\ref{`) : len(m)-1]
	})
	return stem
}

// questionToItem converts a Question to a qti.NewItem.
func questionToItem(q *question.Question) qti.NewItem {
	text := exam.MarkdownToHTML(q.Stem)

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
	case question.FillInTheBlank:
		qtype = qti.FillInMultipleBlanksQuestion
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
		generalFeedback = exam.MarkdownToHTML(q.Explanation)
	}

	var blanks map[string]qti.NewBlank
	if q.Type == question.FillInTheBlank {
		blanks = make(map[string]qti.NewBlank, len(q.Blanks))
		for name, b := range q.Blanks {
			blanks[name] = qti.NewBlank{Answers: b.Answers}
		}
	}

	return qti.NewItem{
		Title:           q.Id,
		Text:            text,
		Type:            qtype,
		Points:          float64(q.Points),
		Choices:         choices,
		GeneralFeedback: generalFeedback,
		Answer:          q.Answer,
		Blanks:          blanks,
	}
}

func init() {
	rootCmd.AddCommand(canvasCmd)
	canvasCmd.Flags().StringVarP(&canvasOutput, "output", "o", "", "output zip file path")
}
