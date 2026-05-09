package syntax

import (
	"strings"
	"testing"
)

func TestHighlightLines_GoSource(t *testing.T) {
	h := NewHighlighter()
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

	joined := strings.Join(result, "\n")
	if !strings.Contains(joined, "\x1b[") {
		t.Error("expected ANSI escape codes in highlighted output")
	}
	// TTY16 emits low-ANSI SGR codes (e.g. \x1b[31m); TTY256 would emit \x1b[38;5;Nm.
	// The presence of "38;5;" indicates the wrong formatter is in use.
	if strings.Contains(joined, "38;5;") {
		t.Error("expected low-ANSI SGR codes, not 256-color codes: output should not contain 38;5;")
	}
}

func TestHighlightLines_UnknownLang(t *testing.T) {
	h := NewHighlighter()
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
	h := NewHighlighter()
	result := h.HighlightLines("main.go", []string{})

	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d elements", len(result))
	}
}

func TestHighlightLines_EmptyContent(t *testing.T) {
	h := NewHighlighter()
	lines := []string{"", ""}
	result := h.HighlightLines("main.go", lines)

	if len(result) != len(lines) {
		t.Fatalf("expected %d lines, got %d", len(lines), len(result))
	}
}
