package question

import "fmt"

// Warning is a validation warning associated with a specific question ID.
type Warning struct {
	Id      string
	Message string
}

// String returns "id: message".
func (w Warning) String() string {
	return fmt.Sprintf("%s: %s", w.Id, w.Message)
}

// ValidateBank performs deep validation on all questions in a bank,
// returning warnings for any issues found.
func ValidateBank(bank Bank) []Warning {
	var warnings []Warning
	for _, item := range bank {
		if q, ok := item.(*Question); ok {
			for _, msg := range q.DeepValidate() {
				warnings = append(warnings, Warning{Id: q.Id, Message: msg})
			}
		}
		if g, ok := item.(*QuestionGroup); ok {
			for _, part := range g.Parts {
				for _, msg := range part.DeepValidate() {
					warnings = append(warnings, Warning{Id: part.Id, Message: msg})
				}
			}
		}
	}
	return warnings
}
