package scantron

import (
	"fmt"

	"github.com/tchajed/question-bank/exam"
	"github.com/tchajed/question-bank/question"
)

// AnswerKey holds the correct 1-based answer index for each question.
// 0 means the question is not gradeable via scantron (short-answer, fill-in-blank).
type AnswerKey []int

// DeriveAnswerKey extracts the correct answer index for each question
// from a resolved exam, using the flat question ordering.
func DeriveAnswerKey(resolved *exam.ResolvedExam) AnswerKey {
	questions := resolved.FlattenQuestions()
	key := make(AnswerKey, len(questions))
	for i, q := range questions {
		key[i] = q.CorrectChoiceIndex()
	}
	return key
}

// NumChoices returns the number of choices for each question in the exam.
func NumChoices(resolved *exam.ResolvedExam) []int {
	questions := resolved.FlattenQuestions()
	result := make([]int, len(questions))
	for i, q := range questions {
		result[i] = len(q.Choices)
	}
	return result
}

// QuestionIDs returns the question ID for each question in flat order.
func QuestionIDs(resolved *exam.ResolvedExam) []string {
	questions := resolved.FlattenQuestions()
	ids := make([]string, len(questions))
	for i, q := range questions {
		ids[i] = q.Id
	}
	return ids
}

// QuestionTypes returns the question type for each question in flat order.
func QuestionTypes(resolved *exam.ResolvedExam) []question.QuestionType {
	questions := resolved.FlattenQuestions()
	types := make([]question.QuestionType, len(questions))
	for i, q := range questions {
		types[i] = q.Type
	}
	return types
}

// GradedRecord augments a StudentRecord with per-question correctness.
type GradedRecord struct {
	Student    *StudentRecord
	Correct    []bool
	NumCorrect int
	NumTotal   int
}

// Grade checks each response against the answer key.
func Grade(record *StudentRecord, key AnswerKey) (*GradedRecord, error) {
	if len(record.Responses) != len(key) {
		return nil, fmt.Errorf("response count %d does not match answer key length %d",
			len(record.Responses), len(key))
	}

	correct := make([]bool, len(key))
	numCorrect := 0
	numTotal := 0
	for i, resp := range record.Responses {
		if key[i] == 0 {
			// Not gradeable (short-answer, etc.)
			continue
		}
		numTotal++
		if resp == key[i] {
			correct[i] = true
			numCorrect++
		}
	}

	return &GradedRecord{
		Student:    record,
		Correct:    correct,
		NumCorrect: numCorrect,
		NumTotal:   numTotal,
	}, nil
}
