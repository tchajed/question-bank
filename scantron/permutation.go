package scantron

import (
	"fmt"

	"github.com/tchajed/question-bank/exam"
)

// Permutation maps question positions from one exam version to another.
// Permutation[i] = j means question at position i in the source version
// corresponds to position j in the canonical version (both 0-indexed).
type Permutation []int

// DerivePermutation computes the permutation that maps a version's question
// order to the canonical question order. Both resolved exams must contain
// exactly the same set of question IDs.
func DerivePermutation(canonical, version *exam.ResolvedExam) (Permutation, error) {
	canonQuestions := canonical.FlattenQuestions()
	versionQuestions := version.FlattenQuestions()

	if len(canonQuestions) != len(versionQuestions) {
		return nil, fmt.Errorf("question count mismatch: canonical has %d, version has %d",
			len(canonQuestions), len(versionQuestions))
	}

	// Build index: question ID → canonical position
	canonIndex := make(map[string]int, len(canonQuestions))
	for i, q := range canonQuestions {
		canonIndex[q.Id] = i
	}

	perm := make(Permutation, len(versionQuestions))
	for i, q := range versionQuestions {
		canonPos, ok := canonIndex[q.Id]
		if !ok {
			return nil, fmt.Errorf("version question %q not found in canonical exam", q.Id)
		}
		perm[i] = canonPos
	}

	return perm, nil
}

// Identity returns the identity permutation for n questions.
func Identity(n int) Permutation {
	perm := make(Permutation, n)
	for i := range perm {
		perm[i] = i
	}
	return perm
}

// Reorder applies the permutation to a response slice, returning a new slice
// in canonical order.
func (p Permutation) Reorder(responses []int) ([]int, error) {
	if len(responses) != len(p) {
		return nil, fmt.Errorf("response length %d does not match permutation length %d",
			len(responses), len(p))
	}
	result := make([]int, len(p))
	for i, canonPos := range p {
		result[canonPos] = responses[i]
	}
	return result, nil
}

// VersionMap associates SpecialCodes values with their permutations.
type VersionMap map[string]Permutation

// ReorderAll applies version-specific permutations to all records in the data.
// Records whose SpecialCodes match a key in versions get that permutation
// applied; all others are treated as canonical (identity permutation).
// Returns a new ScantronData with reordered responses.
func ReorderAll(data *ScantronData, versions VersionMap) (*ScantronData, error) {
	identity := Identity(data.NumQuestions)
	result := &ScantronData{
		NumQuestions: data.NumQuestions,
		Records:      make([]*StudentRecord, len(data.Records)),
	}

	for i, rec := range data.Records {
		perm := identity
		if p, ok := versions[rec.SpecialCodes]; ok {
			perm = p
		}
		reordered, err := perm.Reorder(rec.Responses)
		if err != nil {
			return nil, fmt.Errorf("student %s %s (ID %s): %w",
				rec.FirstName, rec.LastName, rec.ID, err)
		}
		newRec := *rec
		newRec.Responses = reordered
		result.Records[i] = &newRec
	}

	return result, nil
}
