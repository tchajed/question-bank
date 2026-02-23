package exam

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
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
	doc := goldmark.New(goldmark.WithExtensions(extension.Table)).Parser().Parse(reader)

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

	case *extast.Table:
		// Build column spec from header cells (booktabs: no vertical rules).
		colSpec := ""
		if header := n.FirstChild(); header != nil {
			for cell := header.FirstChild(); cell != nil; cell = cell.NextSibling() {
				if tc, ok := cell.(*extast.TableCell); ok {
					switch tc.Alignment {
					case extast.AlignRight:
						colSpec += "r"
					case extast.AlignCenter:
						colSpec += "c"
					default:
						colSpec += "l"
					}
				}
			}
		}
		fmt.Fprintf(buf, "\\begin{tabular}{%s}\n\\toprule\n", colSpec)
		texWalkChildren(buf, src, n)
		buf.WriteString("\\bottomrule\n\\end{tabular}\n\n")

	case *extast.TableHeader:
		writeTableRow(buf, src, n, true)
		buf.WriteString("\\midrule\n")

	case *extast.TableRow:
		writeTableRow(buf, src, n, false)

	case *extast.TableCell:
		texWalkChildren(buf, src, n)

	default:
		// For unhandled node types, recurse into children
		texWalkChildren(buf, src, n)
	}
}

// writeTableRow renders a TableHeader or TableRow node as a LaTeX tabular row.
// If bold is true, cell contents are wrapped in \textbf{}.
func writeTableRow(buf *strings.Builder, src []byte, row ast.Node, bold bool) {
	first := true
	for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
		if !first {
			buf.WriteString(" & ")
		}
		first = false
		if bold {
			buf.WriteString(`\textbf{`)
		}
		texWalkChildren(buf, src, cell)
		if bold {
			buf.WriteString("}")
		}
	}
	buf.WriteString(" \\\\\n")
}

// isLatexLetter reports whether r is a letter that can appear in a LaTeX
// command name (a-z or A-Z).
func isLatexLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// escapeTeX escapes LaTeX special characters in regular text.
// Dollar signs are NOT escaped so that $...$ math passes through unchanged.
// LaTeX commands (\name{arg}[opt]...) are passed through verbatim so that
// constructs like \ref{label} survive the conversion.
func escapeTeX(s string) string {
	var b strings.Builder
	runes := []rune(s)
	i := 0
	for i < len(runes) {
		c := runes[i]
		if c == '\\' && i+1 < len(runes) && isLatexLetter(runes[i+1]) {
			// LaTeX command: pass through \cmdname and any following {arg} or
			// [opt] groups verbatim, handling nesting.
			b.WriteRune(c)
			i++
			for i < len(runes) && isLatexLetter(runes[i]) {
				b.WriteRune(runes[i])
				i++
			}
			for i < len(runes) && (runes[i] == '{' || runes[i] == '[') {
				open, close := runes[i], map[rune]rune{'{': '}', '[': ']'}[runes[i]]
				b.WriteRune(runes[i])
				i++
				depth := 1
				for i < len(runes) && depth > 0 {
					if runes[i] == open {
						depth++
					} else if runes[i] == close {
						depth--
					}
					b.WriteRune(runes[i])
					i++
				}
			}
		} else {
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
			i++
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
