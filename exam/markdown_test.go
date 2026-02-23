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
		{"math caret passthrough", "$x^2$", "$x^2$"},
		{"math underscore passthrough", "$x_i$", "$x_i$"},
		{"math braces passthrough", "$x^{n+1}$", "$x^{n+1}$"},
		{"display math passthrough", "$$x^2$$", "$$x^2$$"},
		{"percent escaped", "50% of", `50\% of`},
		{"code underscore", "`some_var`", `\texttt{some\_var}`},
		{"code dollar", "`$var`", `\texttt{\$var}`},
		{"mixed", "Use **bold** and `code`", `Use \textbf{bold} and \texttt{code}`},
		{"latex command passthrough", `\ref{foo}`, `\ref{foo}`},
		{"latex command with underscore in arg", `\ref{my-group:first}`, `\ref{my-group:first}`},
		{"latex commands in sentence", `see questions \ref{g:first}--\ref{g:last}`, `see questions \ref{g:first}--\ref{g:last}`},
		{"latex command nested braces", `\textbf{a {b} c}`, `\textbf{a {b} c}`},
		{"bare backslash still escaped", `a \ b`, `a \textbackslash{} b`},
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
		{
			"fenced code no language",
			"```\nfoo bar\n```",
			"\\begin{verbatim}\nfoo bar\n\\end{verbatim}",
		},
		{
			"fenced code known language",
			"```c\nint x = 0;\n```",
			"\\begin{lstlisting}[language=C]\nint x = 0;\n\\end{lstlisting}",
		},
		{
			"fenced code go",
			"```go\nfunc main() {}\n```",
			"\\begin{lstlisting}[language=Go]\nfunc main() {}\n\\end{lstlisting}",
		},
		{
			"fenced code unknown language",
			"```rust\nfn main() {}\n```",
			"\\begin{verbatim}\nfn main() {}\n\\end{verbatim}",
		},
		{
			"fenced code python alias",
			"```py\nx = 1\n```",
			"\\begin{lstlisting}[language=Python]\nx = 1\n\\end{lstlisting}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, markdownToTeX(tt.input))
		})
	}
}
