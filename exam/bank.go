package exam

import (
	"slices"

	"github.com/tchajed/question-bank/question"
)

// BankAsExam creates an Exam containing all top-level items in the bank in a
// single unnamed section, sorted by ID. Group parts are omitted because they
// are included via their parent QuestionGroup.
func BankAsExam(bank question.Bank) *Exam {
	var ids []string
	for id, item := range bank {
		switch item.(type) {
		case *question.QuestionGroup:
			ids = append(ids, id)
		case *question.Question:
			if groupIDOfPart(id, bank) == "" {
				ids = append(ids, id)
			}
		}
	}
	slices.Sort(ids)
	return &Exam{
		Title:    "Question bank",
		Sections: []Section{{Questions: ids}},
	}
}
