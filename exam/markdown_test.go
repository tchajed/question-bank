package exam

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkdownToTeX(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain", "plain text", "plain text"},
		{"bold", "**bold**", `\textbf{bold}`},
		{"italic", "*italic*", `\textit{italic}`},
		{"code", "`code`", `\texttt{code}`},
		{"underscore escaped", "some_var", `some\_var`},
		{"dollar passthrough", "$x$", "$x$"},
		{"percent escaped", "50% of", `50\% of`},
		{"code underscore", "`some_var`", `\texttt{some\_var}`},
		{"code dollar", "`$var`", `\texttt{\$var}`},
		{"mixed", "Use **bold** and `code`", `Use \textbf{bold} and \texttt{code}`},
		{"trim whitespace", "\n\ntext\n\n", "text"},
		{
			"multiline stem",
			"First paragraph.\n\nSecond paragraph.",
			"First paragraph.\n\nSecond paragraph.",
		},
		{
			"question with bold",
			"Which of these is **not** an OS benefit?",
			`Which of these is \textbf{not} an OS benefit?`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, markdownToTeX(tt.input))
		})
	}
}
