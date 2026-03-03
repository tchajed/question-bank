// Package scantron handles parsing and processing of scantron result CSVs.
package scantron

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// StudentRecord holds one student's parsed scantron row.
type StudentRecord struct {
	LastName     string
	FirstName    string
	MI           string
	ID           string
	SpecialCodes string
	TotalScore   float64
	TotalPct     float64
	Responses    []int // 1-based answer index per question; 0 = no response
}

// ScantronData holds all parsed records plus metadata.
type ScantronData struct {
	Records      []*StudentRecord
	NumQuestions int
}

// ParseCSV reads a scantron CSV from r. The CSV must have the header format:
// "LastName","FirstName","MI","ID","SpecialCodes","TotalScore","TOTALPCT","_1","_2",...
func ParseCSV(r io.Reader) (*ScantronData, error) {
	cr := csv.NewReader(r)
	header, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}

	// Count question columns (_1, _2, ...)
	numQuestions := 0
	for _, col := range header[7:] {
		if !strings.HasPrefix(col, "_") {
			break
		}
		numQuestions++
	}
	if numQuestions == 0 {
		return nil, fmt.Errorf("no question columns found in header")
	}

	var records []*StudentRecord
	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading row: %w", err)
		}
		if len(row) < 7+numQuestions {
			return nil, fmt.Errorf("row has %d columns, expected at least %d", len(row), 7+numQuestions)
		}

		totalScore, err := strconv.ParseFloat(row[5], 64)
		if err != nil {
			return nil, fmt.Errorf("parsing TotalScore %q: %w", row[5], err)
		}
		totalPct, err := strconv.ParseFloat(row[6], 64)
		if err != nil {
			return nil, fmt.Errorf("parsing TOTALPCT %q: %w", row[6], err)
		}

		responses := make([]int, numQuestions)
		for i := range numQuestions {
			val := strings.TrimSpace(row[7+i])
			if val == "" {
				responses[i] = 0
				continue
			}
			if val == "*" {
				responses[i] = 0
			} else {
				v, err := strconv.Atoi(val)
				if err != nil {
					return nil, fmt.Errorf("parsing response for question %d: %w", i+1, err)
				}
				responses[i] = v
			}
		}

		records = append(records, &StudentRecord{
			LastName:     row[0],
			FirstName:    row[1],
			MI:           row[2],
			ID:           row[3],
			SpecialCodes: strings.TrimSpace(row[4]),
			TotalScore:   totalScore,
			TotalPct:     totalPct,
			Responses:    responses,
		})
	}

	return &ScantronData{
		Records:      records,
		NumQuestions: numQuestions,
	}, nil
}

// WriteCSV writes scantron data in the same CSV format as the input.
func WriteCSV(w io.Writer, data *ScantronData) error {
	cw := csv.NewWriter(w)

	// Header
	header := []string{"LastName", "FirstName", "MI", "ID", "SpecialCodes", "TotalScore", "TOTALPCT"}
	for i := range data.NumQuestions {
		header = append(header, fmt.Sprintf("_%d", i+1))
	}
	if err := cw.Write(header); err != nil {
		return err
	}

	for _, r := range data.Records {
		row := []string{
			r.LastName, r.FirstName, r.MI, r.ID, r.SpecialCodes,
			strconv.FormatFloat(r.TotalScore, 'f', -1, 64),
			strconv.FormatFloat(r.TotalPct, 'f', -1, 64),
		}
		for _, resp := range r.Responses {
			row = append(row, strconv.Itoa(resp))
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}

	cw.Flush()
	return cw.Error()
}
