package qti_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tchajed/question-bank/qti"
)

func TestParseZip(t *testing.T) {
	quizzes, err := qti.ParseZip("testdata/cs537-quiz.zip")
	require.NoError(t, err)
	require.Len(t, quizzes, 1)
	quiz := quizzes[0]

	// Assessment metadata
	assert.Equal(t, "Test quiz", quiz.Meta.Title)
	assert.Equal(t, 3.0, quiz.Meta.PointsPossible)
	assert.Equal(t, 1, quiz.Meta.AllowedAttempts)
	assert.Equal(t, "assignment", quiz.Meta.QuizType)

	// Assessment
	assert.Equal(t, "Test quiz", quiz.Assessment.Root.Title)
	require.Len(t, quiz.Assessment.Root.Items, 3)

	// Item 0: true/false
	item0 := quiz.Assessment.Root.Items[0]
	assert.Equal(t, "Uniprocessor", item0.Title)
	assert.Equal(t, "true_false_question", item0.QuestionType())
	assert.Equal(t, "1.0", item0.PointsPossible())
	require.NotNil(t, item0.Presentation.ResponseLid())
	assert.Equal(t, "Single", item0.Presentation.ResponseLid().RCardinality)
	assert.Len(t, item0.Presentation.ResponseLid().Choices, 2)

	// Item 1: multiple choice (single answer)
	item1 := quiz.Assessment.Root.Items[1]
	assert.Equal(t, "multiple_choice_question", item1.QuestionType())
	require.NotNil(t, item1.Presentation.ResponseLid())
	assert.Equal(t, "Single", item1.Presentation.ResponseLid().RCardinality)
	assert.Len(t, item1.Presentation.ResponseLid().Choices, 3)

	// Item 2: multiple answers
	item2 := quiz.Assessment.Root.Items[2]
	assert.Equal(t, "multiple_answers_question", item2.QuestionType())
	require.NotNil(t, item2.Presentation.ResponseLid())
	assert.Equal(t, "Multiple", item2.Presentation.ResponseLid().RCardinality)
	assert.Len(t, item2.Presentation.ResponseLid().Choices, 3)
}

func TestItemFeedback(t *testing.T) {
	quizzes, err := qti.ParseZip("testdata/cs537-quiz.zip")
	require.NoError(t, err)
	require.Len(t, quizzes, 1)
	quiz := quizzes[0]

	// Item 0 has 3 feedback entries
	item0 := quiz.Assessment.Root.Items[0]
	assert.Len(t, item0.ItemFeedback, 3)
	feedbackMap := make(map[string]string)
	for _, fb := range item0.ItemFeedback {
		feedbackMap[fb.Ident] = fb.Material.MatText.Text
	}
	assert.Contains(t, feedbackMap, "general_fb")
	assert.Contains(t, feedbackMap, "correct_fb")
	assert.Contains(t, feedbackMap, "general_incorrect_fb")
}

// cs537Quiz returns a NewQuiz matching testdata/cs537-quiz.zip.
func cs537Quiz() *qti.NewQuiz {
	return &qti.NewQuiz{
		ID:              "g06a00cfb4ed595904b671c0c2d6562d5",
		Title:           "Test quiz",
		Description:     "<p>This quiz is a test of Canvas importing</p>",
		PointsPossible:  3.0,
		AllowedAttempts: 1,
		QuizType:        "assignment",
		Items: []qti.NewItem{
			{
				ID:    "g074cef231ce6afa1ab211121f16dc838",
				Title: "Uniprocessor",
				Type:  qti.TrueFalseQuestion,
				Text:  `<div><p><span>In a uniprocessor system, there can be more than one processes in the READY and BLOCKED states and at most one process in RUNNING state.</span></p></div>`,
				Points: 1.0,
				Choices: []qti.NewChoice{
					{ID: "8205", Text: "True", Correct: true},
					{ID: "2854", Text: "False"},
				},
				GeneralFeedback:   `<p>A uniprocessor system can have no processes running (if the scheduler is still working).</p>`,
				CorrectFeedback:   `<p>Good job!</p>`,
				IncorrectFeedback: `<p>Not quite: a uniprocessor system cannot RUN more than one process.</p>`,
			},
			{
				ID:    "geac9ad5e94b1264a4988a21c4a5b0022",
				Title: "Question",
				Type:  qti.MultipleChoiceQuestion,
				Text:  `<div><p class="p1">Which of these is <strong>not</strong> an application benefit of an operating system?</p></div>`,
				Points: 1.0,
				Choices: []qti.NewChoice{
					{ID: "8321", Text: "A set of simpler abstractions against which to program"},
					{ID: "6847", Text: "Independence from specific hardware and devices"},
					{ID: "5131", Text: "More control over how hardware is used", Correct: true},
				},
				GeneralFeedback: `<p class="p1">An operating system gives less direct control over the hardware to applications.</p>`,
			},
			{
				ID:    "g0f150c14d07dfa656046e19a7372f0db",
				Title: "Multiple",
				Type:  qti.MultipleAnswersQuestion,
				Text:  `<div><p>You should select A and C for this question.</p></div>`,
				Points: 1.0,
				Choices: []qti.NewChoice{
					{ID: "6295", Text: "A", Correct: true},
					{ID: "1153", Text: "B"},
					{ID: "4634", Text: "C", Correct: true},
				},
				IncorrectFeedback: `<p>You did not follow the instructions</p>`,
			},
		},
	}
}

func TestWriteZip(t *testing.T) {
	f, err := os.CreateTemp("", "quiz-*.zip")
	require.NoError(t, err)
	f.Close()
	defer os.Remove(f.Name())

	err = qti.WriteZip(f.Name(), cs537Quiz())
	require.NoError(t, err)

	quizzes, err := qti.ParseZip(f.Name())
	require.NoError(t, err)
	require.Len(t, quizzes, 1)
	quiz := quizzes[0]

	// Metadata
	assert.Equal(t, "Test quiz", quiz.Meta.Title)
	assert.Equal(t, "<p>This quiz is a test of Canvas importing</p>", quiz.Meta.Description)
	assert.Equal(t, 3.0, quiz.Meta.PointsPossible)
	assert.Equal(t, 1, quiz.Meta.AllowedAttempts)
	assert.Equal(t, "assignment", quiz.Meta.QuizType)

	// Assessment
	assert.Equal(t, "Test quiz", quiz.Assessment.Root.Title)
	require.Len(t, quiz.Assessment.Root.Items, 3)

	// Item 0: true/false
	item0 := quiz.Assessment.Root.Items[0]
	assert.Equal(t, "Uniprocessor", item0.Title)
	assert.Equal(t, "true_false_question", item0.QuestionType())
	assert.Equal(t, "1.0", item0.PointsPossible())
	require.NotNil(t, item0.Presentation.ResponseLid())
	assert.Equal(t, "Single", item0.Presentation.ResponseLid().RCardinality)
	require.Len(t, item0.Presentation.ResponseLid().Choices, 2)
	assert.Equal(t, "True", item0.Presentation.ResponseLid().Choices[0].Material.MatText.Text)
	assert.Equal(t, "False", item0.Presentation.ResponseLid().Choices[1].Material.MatText.Text)

	// Item 0: correct answer is "8205" (True)
	var correctID string
	for _, cond := range item0.ResProcessing.RespConditions {
		if cond.SetVar != nil && cond.SetVar.Value == "100" {
			if len(cond.ConditionVar.VarEquals) == 1 {
				correctID = cond.ConditionVar.VarEquals[0].Value
			}
		}
	}
	assert.Equal(t, "8205", correctID)

	// Item 0: all three feedback slots present
	feedbackMap := make(map[string]string)
	for _, fb := range item0.ItemFeedback {
		feedbackMap[fb.Ident] = fb.Material.MatText.Text
	}
	assert.Contains(t, feedbackMap, "general_fb")
	assert.Contains(t, feedbackMap, "correct_fb")
	assert.Contains(t, feedbackMap, "general_incorrect_fb")

	// Item 1: multiple choice
	item1 := quiz.Assessment.Root.Items[1]
	assert.Equal(t, "Question", item1.Title)
	assert.Equal(t, "multiple_choice_question", item1.QuestionType())
	assert.Equal(t, "Single", item1.Presentation.ResponseLid().RCardinality)
	require.Len(t, item1.Presentation.ResponseLid().Choices, 3)
	// Only general feedback
	assert.Len(t, item1.ItemFeedback, 1)
	assert.Equal(t, "general_fb", item1.ItemFeedback[0].Ident)

	// Item 1: correct answer is "5131" (More control...)
	correctID = ""
	for _, cond := range item1.ResProcessing.RespConditions {
		if cond.SetVar != nil && cond.SetVar.Value == "100" {
			if len(cond.ConditionVar.VarEquals) == 1 {
				correctID = cond.ConditionVar.VarEquals[0].Value
			}
		}
	}
	assert.Equal(t, "5131", correctID)

	// Item 2: multiple answers
	item2 := quiz.Assessment.Root.Items[2]
	assert.Equal(t, "Multiple", item2.Title)
	assert.Equal(t, "multiple_answers_question", item2.QuestionType())
	assert.Equal(t, "Multiple", item2.Presentation.ResponseLid().RCardinality)
	require.Len(t, item2.Presentation.ResponseLid().Choices, 3)

	// Item 2: AND condition with 2 correct choices and 1 NOT
	var correctCond *qti.RespCondition
	for i := range item2.ResProcessing.RespConditions {
		c := &item2.ResProcessing.RespConditions[i]
		if c.SetVar != nil && c.SetVar.Value == "100" {
			correctCond = c
			break
		}
	}
	require.NotNil(t, correctCond)
	require.NotNil(t, correctCond.ConditionVar.And)
	assert.Len(t, correctCond.ConditionVar.And.VarEquals, 2)
	assert.Len(t, correctCond.ConditionVar.And.Nots, 1)
	// Correct IDs are 6295 and 4634; incorrect is 1153
	correctIDs := []string{
		correctCond.ConditionVar.And.VarEquals[0].Value,
		correctCond.ConditionVar.And.VarEquals[1].Value,
	}
	assert.ElementsMatch(t, []string{"6295", "4634"}, correctIDs)
	assert.Equal(t, "1153", correctCond.ConditionVar.And.Nots[0].VarEquals[0].Value)
}

func TestCorrectAnswers(t *testing.T) {
	quizzes, err := qti.ParseZip("testdata/cs537-quiz.zip")
	require.NoError(t, err)
	require.Len(t, quizzes, 1)
	quiz := quizzes[0]

	// Item 0 (true/false): correct answer is "8205" (True)
	item0 := quiz.Assessment.Root.Items[0]
	var correctIdent string
	for _, cond := range item0.ResProcessing.RespConditions {
		if cond.SetVar != nil && cond.SetVar.Value == "100" {
			if len(cond.ConditionVar.VarEquals) == 1 {
				correctIdent = cond.ConditionVar.VarEquals[0].Value
			}
		}
	}
	assert.Equal(t, "8205", correctIdent)

	// Item 2 (multiple answers): correct answers are 6295 and 4634 (A and C), not 1153 (B)
	item2 := quiz.Assessment.Root.Items[2]
	var correctCond *qti.RespCondition
	for i := range item2.ResProcessing.RespConditions {
		c := &item2.ResProcessing.RespConditions[i]
		if c.SetVar != nil && c.SetVar.Value == "100" {
			correctCond = c
			break
		}
	}
	require.NotNil(t, correctCond)
	require.NotNil(t, correctCond.ConditionVar.And)
	assert.Len(t, correctCond.ConditionVar.And.VarEquals, 2)
	assert.Len(t, correctCond.ConditionVar.And.Nots, 1)
}

func TestWriteZipFillInBlanks(t *testing.T) {
	f, err := os.CreateTemp("", "quiz-blanks-*.zip")
	require.NoError(t, err)
	f.Close()
	defer os.Remove(f.Name())

	quiz := &qti.NewQuiz{
		Title:          "Blanks quiz",
		PointsPossible: 1.0,
		Items: []qti.NewItem{
			{
				Title:  "FIB",
				Type:   qti.FillInMultipleBlanksQuestion,
				Text:   "<p>A [lock_type] ensures only [n] thread(s).</p>",
				Points: 1.0,
				Blanks: map[string]qti.NewBlank{
					"lock_type": {Answers: []string{"mutex"}},
					"n":         {Answers: []string{"1", "one"}},
				},
			},
		},
	}

	err = qti.WriteZip(f.Name(), quiz)
	require.NoError(t, err)

	quizzes, err := qti.ParseZip(f.Name())
	require.NoError(t, err)
	require.Len(t, quizzes, 1)
	parsed := quizzes[0]

	assert.Equal(t, "Blanks quiz", parsed.Meta.Title)
	require.Len(t, parsed.Assessment.Root.Items, 1)

	item := parsed.Assessment.Root.Items[0]
	assert.Equal(t, "fill_in_multiple_blanks_question", item.QuestionType())
	// Should have 2 response_lids (one per blank)
	require.Len(t, item.Presentation.ResponseLids, 2)
	// Sorted: lock_type, n
	assert.Equal(t, "response_lock_type", item.Presentation.ResponseLids[0].Ident)
	assert.Equal(t, "response_n", item.Presentation.ResponseLids[1].Ident)
	assert.Len(t, item.Presentation.ResponseLids[0].Choices, 1)
	assert.Len(t, item.Presentation.ResponseLids[1].Choices, 2)

	// Scoring: 2 conditions with Add action
	var addConds int
	for _, c := range item.ResProcessing.RespConditions {
		if c.SetVar != nil && c.SetVar.Action == "Add" {
			addConds++
			assert.Equal(t, "50.00", c.SetVar.Value)
		}
	}
	assert.Equal(t, 2, addConds)
}

func TestWriteZipMultiple(t *testing.T) {
	f, err := os.CreateTemp("", "quiz-multi-*.zip")
	require.NoError(t, err)
	f.Close()
	defer os.Remove(f.Name())

	quiz1 := cs537Quiz()
	quiz2 := &qti.NewQuiz{
		Title:           "Second quiz",
		PointsPossible:  1.0,
		AllowedAttempts: 1,
		QuizType:        "assignment",
		Items: []qti.NewItem{
			{
				Title:  "Q1",
				Type:   qti.TrueFalseQuestion,
				Points: 1.0,
				Choices: []qti.NewChoice{
					{Text: "True", Correct: true},
					{Text: "False"},
				},
			},
		},
	}

	err = qti.WriteZip(f.Name(), quiz1, quiz2)
	require.NoError(t, err)

	quizzes, err := qti.ParseZip(f.Name())
	require.NoError(t, err)
	require.Len(t, quizzes, 2)

	assert.Equal(t, "Test quiz", quizzes[0].Meta.Title)
	assert.Len(t, quizzes[0].Assessment.Root.Items, 3)

	assert.Equal(t, "Second quiz", quizzes[1].Meta.Title)
	assert.Len(t, quizzes[1].Assessment.Root.Items, 1)
}
