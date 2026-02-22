package exam

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// markdownToTeX converts a Markdown string to LaTeX.
//
// Handles inline formatting (bold → \textbf, italic → \textit, code →
// \texttt), paragraphs, lists, and fenced code blocks. Dollar signs in regular
// text are NOT escaped so that $...$ math passes through unchanged.
func markdownToTeX(md string) string {
	if md == "" {
		return ""
	}
	src := []byte(strings.TrimSpace(md))
	reader := text.NewReader(src)
	doc := goldmark.New().Parser().Parse(reader)

	var buf strings.Builder
	texWalk(&buf, src, doc)
	return strings.TrimSpace(buf.String())
}

func texWalkChildren(buf *strings.Builder, src []byte, node ast.Node) {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		texWalk(buf, src, child)
	}
}

func texWalk(buf *strings.Builder, src []byte, node ast.Node) {
	switch n := node.(type) {
	case *ast.Document:
		texWalkChildren(buf, src, n)

	case *ast.Paragraph, *ast.TextBlock:
		texWalkChildren(buf, src, n)
		buf.WriteString("\n\n")

	case *ast.Heading:
		buf.WriteString(`\textbf{`)
		texWalkChildren(buf, src, n)
		buf.WriteString("}\n\n")

	case *ast.Text:
		buf.WriteString(escapeTeX(string(n.Segment.Value(src))))
		if n.HardLineBreak() {
			buf.WriteString("\\\\\n")
		} else if n.SoftLineBreak() {
			buf.WriteString("\n")
		}

	case *ast.String:
		buf.WriteString(escapeTeX(string(n.Value)))

	case *ast.Emphasis:
		if n.Level == 2 {
			buf.WriteString(`\textbf{`)
		} else {
			buf.WriteString(`\textit{`)
		}
		texWalkChildren(buf, src, n)
		buf.WriteString("}")

	case *ast.CodeSpan:
		buf.WriteString(`\texttt{`)
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			switch t := child.(type) {
			case *ast.Text:
				buf.WriteString(escapeCodeTeX(string(t.Segment.Value(src))))
			case *ast.String:
				buf.WriteString(escapeCodeTeX(string(t.Value)))
			}
		}
		buf.WriteString("}")

	case *ast.FencedCodeBlock:
		buf.WriteString("\\begin{verbatim}\n")
		for i := 0; i < n.Lines().Len(); i++ {
			line := n.Lines().At(i)
			buf.WriteString(string(line.Value(src)))
		}
		buf.WriteString("\\end{verbatim}\n")

	case *ast.CodeBlock:
		buf.WriteString("\\begin{verbatim}\n")
		for i := 0; i < n.Lines().Len(); i++ {
			line := n.Lines().At(i)
			buf.WriteString(string(line.Value(src)))
		}
		buf.WriteString("\\end{verbatim}\n")

	case *ast.List:
		env := "itemize"
		if n.IsOrdered() {
			env = "enumerate"
		}
		buf.WriteString("\\begin{" + env + "}\n")
		texWalkChildren(buf, src, n)
		buf.WriteString("\\end{" + env + "}\n")

	case *ast.ListItem:
		buf.WriteString("\\item ")
		texWalkChildren(buf, src, n)

	case *ast.Link:
		// Render link text only (URL is not useful in print)
		texWalkChildren(buf, src, n)

	default:
		// For unhandled node types, recurse into children
		texWalkChildren(buf, src, n)
	}
}

// escapeTeX escapes LaTeX special characters in regular text.
// Dollar signs are NOT escaped so that $...$ math passes through unchanged.
func escapeTeX(s string) string {
	var b strings.Builder
	for _, c := range s {
		switch c {
		case '\\':
			b.WriteString(`\textbackslash{}`)
		case '&':
			b.WriteString(`\&`)
		case '%':
			b.WriteString(`\%`)
		case '#':
			b.WriteString(`\#`)
		case '_':
			b.WriteString(`\_`)
		case '{':
			b.WriteString(`\{`)
		case '}':
			b.WriteString(`\}`)
		case '~':
			b.WriteString(`\~{}`)
		case '^':
			b.WriteString(`\^{}`)
		default:
			b.WriteRune(c)
		}
	}
	return b.String()
}

// escapeCodeTeX escapes all LaTeX special characters including $ for code contexts.
func escapeCodeTeX(s string) string {
	var b strings.Builder
	for _, c := range s {
		switch c {
		case '\\':
			b.WriteString(`\textbackslash{}`)
		case '&':
			b.WriteString(`\&`)
		case '%':
			b.WriteString(`\%`)
		case '$':
			b.WriteString(`\$`)
		case '#':
			b.WriteString(`\#`)
		case '_':
			b.WriteString(`\_`)
		case '{':
			b.WriteString(`\{`)
		case '}':
			b.WriteString(`\}`)
		case '~':
			b.WriteString(`\~{}`)
		case '^':
			b.WriteString(`\^{}`)
		default:
			b.WriteRune(c)
		}
	}
	return b.String()
}
