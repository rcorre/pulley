package syntax

import (
	"strings"
	"testing"
)

func TestHighlightLines_GoSource(t *testing.T) {
	h := NewHighlighter("")
	lines := []string{
		"package main",
		"",
		"import \"fmt\"",
		"",
		"func main() {",
		"\tfmt.Println(\"hello\")",
		"}",
	}

	result := h.HighlightLines("main.go", lines)

	if len(result) != len(lines) {
		t.Fatalf("expected %d lines, got %d", len(lines), len(result))
	}

	hadAnsi := false
	for _, line := range result {
		if strings.ContainsAny(line, "\x1b") {
			hadAnsi = true
			break
		}
	}

	if !hadAnsi {
		t.Error("expected at least some lines to contain ANSI escape codes")
	}
}

func TestHighlightLines_UnknownLang(t *testing.T) {
	h := NewHighlighter("")
	lines := []string{
		"some random content",
		"that is not code",
	}

	result := h.HighlightLines("unknown.xyz", lines)

	if len(result) != len(lines) {
		t.Fatalf("expected %d lines, got %d", len(lines), len(result))
	}

	for i, line := range result {
		if line != lines[i] {
			t.Errorf("line %d: expected unchanged content, got %q", i, line)
		}
	}
}

func TestHighlightLines_EmptyInput(t *testing.T) {
	h := NewHighlighter("")
	result := h.HighlightLines("main.go", []string{})

	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d elements", len(result))
	}
}

func TestHighlightLines_EmptyContent(t *testing.T) {
	h := NewHighlighter("")
	lines := []string{"", ""}
	result := h.HighlightLines("main.go", lines)

	if len(result) != len(lines) {
		t.Fatalf("expected %d lines, got %d", len(lines), len(result))
	}
}
