// Package syntax provides syntax highlighting using chroma with base16 ANSI colors.
package syntax

import (
	"bytes"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// Highlighter provides syntax highlighting using base16-compatible indexed ANSI colors.
type Highlighter struct {
	style     *chroma.Style
	formatter chroma.Formatter
}

// NewHighlighter creates a Highlighter with base16 ANSI color mappings.
func NewHighlighter() *Highlighter {
	return &Highlighter{
		style:     base16Style(),
		formatter: formatters.TTY16,
	}
}

// HighlightLines syntax highlights the given lines using the filename to detect language.
// Returns the lines with ANSI escape codes for coloring.
// If the language cannot be determined, returns the lines unchanged.
func (h *Highlighter) HighlightLines(filename string, lines []string) []string {
	if len(lines) == 0 {
		return lines
	}

	lexer := lexers.Match(filename)
	if lexer == nil {
		return lines
	}

	lexer = chroma.Coalesce(lexer)

	content := strings.Join(lines, "\n")
	if content == "" {
		return lines
	}

	it, err := lexer.Tokenise(nil, content)
	if err != nil {
		return lines
	}

	var buf bytes.Buffer
	if err := h.formatter.Format(&buf, h.style, it); err != nil {
		return lines
	}

	highlighted := buf.String()
	result := strings.Split(highlighted, "\n")
	if len(result) > 0 && result[len(result)-1] == "" {
		result = result[:len(result)-1]
	}

	if len(result) != len(lines) {
		return lines
	}

	return result
}

// base16Style creates a chroma Style using base16 ANSI named colors.
// Uses #ansiblack through #ansiwhite for base16 palette compatibility.
func base16Style() *chroma.Style {
	return styles.Register(chroma.MustNewStyle("base16", chroma.StyleEntries{
		chroma.Text:                   "#ansilightgray",
		chroma.Whitespace:             "#ansidarkgray",
		chroma.Error:                  "#ansired",
		chroma.Comment:                "#ansibrown",
		chroma.CommentPreproc:         "#ansibrown",
		chroma.Keyword:                "#ansipurple",
		chroma.KeywordConstant:        "#ansipurple",
		chroma.KeywordDeclaration:     "#ansipurple",
		chroma.KeywordNamespace:       "#ansipurple",
		chroma.KeywordPseudo:          "#ansipurple",
		chroma.KeywordReserved:        "#ansipurple",
		chroma.KeywordType:            "#ansipurple",
		chroma.Operator:               "#ansilightgray",
		chroma.OperatorWord:           "#ansipurple",
		chroma.Punctuation:            "#ansilightgray",
		chroma.Name:                   "#ansilightgray",
		chroma.NameAttribute:          "#ansilightgray",
		chroma.NameBuiltin:            "#ansidarkgray",
		chroma.NameBuiltinPseudo:      "#ansidarkgray",
		chroma.NameClass:              "#ansidarkblue",
		chroma.NameConstant:           "#ansidarkblue",
		chroma.NameDecorator:          "#ansidarkblue",
		chroma.NameEntity:             "#ansidarkblue",
		chroma.NameException:          "#ansidarkblue",
		chroma.NameFunction:           "#ansiblue",
		chroma.NameLabel:              "#ansilightgray",
		chroma.NameNamespace:          "#ansidarkblue",
		chroma.NameOther:              "#ansilightgray",
		chroma.NameProperty:           "#ansilightgray",
		chroma.NameTag:                "#ansidarkblue",
		chroma.NameVariable:           "#ansilightgray",
		chroma.NameVariableClass:      "#ansilightgray",
		chroma.NameVariableGlobal:     "#ansilightgray",
		chroma.NameVariableInstance:   "#ansilightgray",
		chroma.Literal:                "#ansigreen",
		chroma.LiteralDate:            "#ansigreen",
		chroma.LiteralString:          "#ansigreen",
		chroma.LiteralStringAffix:     "#ansigreen",
		chroma.LiteralStringBacktick:  "#ansigreen",
		chroma.LiteralStringChar:      "#ansigreen",
		chroma.LiteralStringDelimiter: "#ansigreen",
		chroma.LiteralStringDoc:       "#ansigreen",
		chroma.LiteralStringDouble:    "#ansigreen",
		chroma.LiteralStringEscape:    "#ansigreen",
		chroma.LiteralStringHeredoc:   "#ansigreen",
		chroma.LiteralStringInterpol:  "#ansigreen",
		chroma.LiteralStringOther:     "#ansigreen",
		chroma.LiteralStringRegex:     "#ansigreen",
		chroma.LiteralStringSingle:    "#ansigreen",
		chroma.LiteralStringSymbol:    "#ansigreen",
		chroma.LiteralNumber:          "#ansidarkblue",
		chroma.LiteralNumberBin:       "#ansidarkblue",
		chroma.LiteralNumberFloat:     "#ansidarkblue",
		chroma.LiteralNumberHex:       "#ansidarkblue",
		chroma.LiteralNumberInteger:   "#ansidarkblue",
		chroma.LiteralNumberOct:       "#ansidarkblue",
		chroma.Generic:                "#ansilightgray",
		chroma.GenericDeleted:         "#ansired",
		chroma.GenericEmph:            "#ansilightgray",
		chroma.GenericError:           "#ansired",
		chroma.GenericHeading:         "#ansipurple",
		chroma.GenericInserted:        "#ansigreen",
		chroma.GenericOutput:          "#ansilightgray",
		chroma.GenericPrompt:          "#ansilightgray",
		chroma.GenericStrong:          "#ansilightgray",
		chroma.GenericSubheading:      "#ansipurple",
		chroma.GenericTraceback:       "#ansired",
		chroma.GenericUnderline:       "#ansilightgray",
	}))
}
