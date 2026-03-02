package cmd

import (
	"bytes"
	"testing"
)

func TestAnswerKey(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "letter",
			args: []string{"answer-key", "-b", "../testdata/bank", "../testdata/exams/exam.toml"},
			want: `question,answer
1,C
2,A
3,B
4,fork()
5,mutex; 1
6,A
7,B
`,
		},
		{
			name: "numeric",
			args: []string{"answer-key", "--numeric", "-b", "../testdata/bank", "../testdata/exams/exam.toml"},
			want: `question,answer
_1,3
_2,1
_3,2
_4,fork()
_5,mutex; 1
_6,1
_7,2
`,
		},
		{
			name: "row",
			args: []string{"answer-key", "--row", "-b", "../testdata/bank", "../testdata/exams/exam.toml"},
			want: `1,2,3,4,5,6,7
C,A,B,fork(),mutex; 1,A,B
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetArgs(tt.args)
			// Reset flags to defaults before each run
			numericAnswerKey = false
			rowAnswerKey = false
			if err := rootCmd.Execute(); err != nil {
				t.Fatal(err)
			}
			got := buf.String()
			if got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}
