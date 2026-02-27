package exam_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tchajed/question-bank/exam"
	"github.com/tchajed/question-bank/question"
)

func TestBankAsExam(t *testing.T) {
	bank, err := question.LoadBank("../testdata/bank")
	require.NoError(t, err)

	e := exam.BankAsExam(bank)
	resolved, err := e.Resolve(bank)
	require.NoError(t, err)

	// os-001, os-002, processes-001, processes-group-001 (with 2 parts), threads-001, threads-002, vm-001, vm-002, vm-003, vm-004
	assert.Len(t, resolved.Sections[0].Items, 10)
}
