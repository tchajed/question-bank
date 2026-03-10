package scantron

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
)

// QuestionStats holds per-question analysis data.
type QuestionStats struct {
	QuestionNum       int              `json:"question_num"`
	QuestionID        string           `json:"question_id"`
	NumChoices        int              `json:"num_choices"`
	OverallPctCorrect float64          `json:"overall_pct_correct"`
	ByQuintile        [5]QuintileStats `json:"by_quintile"` // index 0=bottom, 4=top
}

// QuintileStats holds statistics for one quintile of students.
type QuintileStats struct {
	N            int       `json:"n"`
	PctCorrect   float64   `json:"pct_correct"`
	ResponseDist []float64 `json:"response_dist"` // fraction choosing each choice (0-indexed); last element = no response
}

// assignQuintiles divides n students (already sorted by score, ascending)
// into 5 quintile groups. Returns a slice of length n where each element
// is the quintile index (0=bottom, 4=top).
func assignQuintiles(n int) []int {
	assignments := make([]int, n)
	if n == 0 {
		return assignments
	}
	for i := range n {
		// Map each student index to a quintile.
		// Quintile q gets students from index q*n/5 to (q+1)*n/5 - 1.
		q := i * 5 / n
		if q >= 5 {
			q = 4
		}
		assignments[i] = q
	}
	return assignments
}

// ItemAnalysis computes per-question statistics grouped by score quintile.
// Students are sorted by TotalScore and divided into 5 equal groups.
func ItemAnalysis(records []*StudentRecord, key AnswerKey, questionIDs []string, numChoices []int) ([]QuestionStats, error) {
	numQ := len(key)
	if len(questionIDs) != numQ {
		return nil, fmt.Errorf("questionIDs length %d does not match answer key length %d", len(questionIDs), numQ)
	}
	if len(numChoices) != numQ {
		return nil, fmt.Errorf("numChoices length %d does not match answer key length %d", len(numChoices), numQ)
	}
	for _, r := range records {
		if len(r.Responses) != numQ {
			return nil, fmt.Errorf("student %s has %d responses, expected %d", r.ID, len(r.Responses), numQ)
		}
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no student records provided")
	}

	// Sort students by TotalScore (ascending) for quintile assignment.
	sorted := make([]*StudentRecord, len(records))
	copy(sorted, records)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].TotalScore < sorted[j].TotalScore
	})

	quintileAssign := assignQuintiles(len(sorted))

	// Count students per quintile.
	quintileN := [5]int{}
	for _, q := range quintileAssign {
		quintileN[q]++
	}

	stats := make([]QuestionStats, numQ)
	for qi := range numQ {
		nc := numChoices[qi]
		stats[qi] = QuestionStats{
			QuestionNum: qi + 1,
			QuestionID:  questionIDs[qi],
			NumChoices:  nc,
		}

		// Initialize quintile stats.
		for q := range 5 {
			stats[qi].ByQuintile[q] = QuintileStats{
				N:            quintileN[q],
				ResponseDist: make([]float64, nc+1), // nc choices + 1 for no response
			}
		}

		// Tally responses and correctness.
		overallCorrect := 0
		quintileCorrect := [5]int{}
		quintileResponseCounts := [5][]int{}
		for q := range 5 {
			quintileResponseCounts[q] = make([]int, nc+1)
		}

		for si, student := range sorted {
			q := quintileAssign[si]
			resp := student.Responses[qi]
			correct := key[qi]

			if resp == correct && correct != 0 {
				overallCorrect++
				quintileCorrect[q]++
			}

			if resp == 0 {
				// No response: last bucket.
				quintileResponseCounts[q][nc]++
			} else if resp >= 1 && resp <= nc {
				// 1-based response to 0-based index.
				quintileResponseCounts[q][resp-1]++
			}
			// Responses outside valid range are ignored.
		}

		stats[qi].OverallPctCorrect = float64(overallCorrect) / float64(len(sorted)) * 100.0

		for q := range 5 {
			if quintileN[q] > 0 {
				stats[qi].ByQuintile[q].PctCorrect = float64(quintileCorrect[q]) / float64(quintileN[q]) * 100.0
				for c := range nc + 1 {
					stats[qi].ByQuintile[q].ResponseDist[c] = float64(quintileResponseCounts[q][c]) / float64(quintileN[q])
				}
			}
		}
	}

	return stats, nil
}

// WriteAnalysisCSV writes item analysis stats as a CSV with columns:
// Question, ID, OverallPctCorrect, Q1_PctCorrect, Q2_PctCorrect, ..., Q5_PctCorrect
func WriteAnalysisCSV(w io.Writer, stats []QuestionStats) error {
	cw := csv.NewWriter(w)

	header := []string{"Question", "ID", "OverallPctCorrect",
		"Q1_PctCorrect", "Q2_PctCorrect", "Q3_PctCorrect", "Q4_PctCorrect", "Q5_PctCorrect"}
	if err := cw.Write(header); err != nil {
		return err
	}

	for _, s := range stats {
		row := []string{
			strconv.Itoa(s.QuestionNum),
			s.QuestionID,
			strconv.FormatFloat(s.OverallPctCorrect, 'f', 1, 64),
		}
		for q := range 5 {
			row = append(row, strconv.FormatFloat(s.ByQuintile[q].PctCorrect, 'f', 1, 64))
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}

	cw.Flush()
	return cw.Error()
}

// WriteAnalysisJSON writes the full item analysis stats as JSON.
func WriteAnalysisJSON(w io.Writer, stats []QuestionStats) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(stats)
}
