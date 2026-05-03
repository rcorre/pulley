package comment

import (
	"strings"
	"testing"

	"github.com/rcorre/pulley/internal/diff"
	"github.com/stretchr/testify/assert"
)

func TestBuildTemplateComment(t *testing.T) {
	line := diff.Line{Kind: diff.LineContext, Content: "return nil", NewLine: 42, DiffPosition: 5}
	tmpl := BuildTemplate("foo.go", line, false)

	assert.Contains(t, tmpl, "# Comment on: foo.go")
	assert.Contains(t, tmpl, "# Line 42: return nil")
	assert.Contains(t, tmpl, "Lines above the blank line will be removed")
	assert.NotContains(t, tmpl, "```suggestion")
	// blank line separates header from body
	assert.Contains(t, tmpl, "\n\n")
}

func TestBuildTemplateSuggestion(t *testing.T) {
	line := diff.Line{Kind: diff.LineAdd, Content: "\tvar x = 1", NewLine: 10, DiffPosition: 3}
	tmpl := BuildTemplate("bar.go", line, true)

	assert.Contains(t, tmpl, "# Suggestion for: bar.go")
	assert.Contains(t, tmpl, "# Line 10: \tvar x = 1")
	assert.Contains(t, tmpl, "```suggestion")
	assert.Contains(t, tmpl, "\tvar x = 1")
	// fence should close
	lines := strings.Split(tmpl, "\n")
	var fences []string
	for _, l := range lines {
		if strings.HasPrefix(l, "```") {
			fences = append(fences, l)
		}
	}
	assert.Len(t, fences, 2, "suggestion template should have opening and closing fences")
}

func TestBuildTemplateOldLineOnly(t *testing.T) {
	line := diff.Line{Kind: diff.LineRemove, Content: "old code", OldLine: 7, DiffPosition: 2}
	tmpl := BuildTemplate("baz.go", line, false)

	assert.Contains(t, tmpl, "# Line 7: old code")
}

func TestParseTemplateStripsHeader(t *testing.T) {
	input := "# this is a header\n# another header\n\nActual comment body.\n"
	got := ParseTemplate(input)
	assert.Equal(t, "Actual comment body.", got)
}

func TestParseTemplatePreservesHashInBody(t *testing.T) {
	input := "# template header\n#\n\n# My Heading\n\nSome text.\n"
	got := ParseTemplate(input)
	assert.Equal(t, "# My Heading\n\nSome text.", got)
}

func TestParseTemplateKeepsSuggestionFence(t *testing.T) {
	input := "# header\n#\n\n```suggestion\nnew code\n```\n"
	got := ParseTemplate(input)
	assert.Equal(t, "```suggestion\nnew code\n```", got)
}

func TestParseTemplateEmptyBody(t *testing.T) {
	// No blank line separator: falls back to full content (trimmed)
	input := "# header\n# another header\n"
	got := ParseTemplate(input)
	// No separator found and no non-# content: full trimmed content
	assert.Equal(t, "# header\n# another header", got)
}

func TestParseTemplateEmptyBodyWithSeparator(t *testing.T) {
	input := "# header\n#\n\n"
	got := ParseTemplate(input)
	assert.Empty(t, got)
}

func TestParseTemplateTrimsSurroundingWhitespace(t *testing.T) {
	input := "# header\n#\n\n  comment with spaces  \n\n"
	got := ParseTemplate(input)
	assert.Equal(t, "comment with spaces", got)
}
