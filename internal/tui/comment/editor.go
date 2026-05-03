// Package comment handles spawning $EDITOR for drafting review comments.
package comment

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rcorre/pulley/internal/diff"
	"github.com/rcorre/pulley/internal/github"
)

// DraftAddedMsg is sent when the user saves a non-empty comment from the editor.
type DraftAddedMsg struct {
	Draft github.DraftComment
}

// BuildTemplate generates the temp file content for the editor.
// '#'-prefixed header lines are stripped by ParseTemplate on save.
func BuildTemplate(filename string, line diff.Line, suggestion bool) string {
	var b strings.Builder

	if suggestion {
		fmt.Fprintf(&b, "# Suggestion for: %s\n", filename)
	} else {
		fmt.Fprintf(&b, "# Comment on: %s\n", filename)
	}

	lineNum := line.NewLine
	if lineNum == 0 {
		lineNum = line.OldLine
	}
	if lineNum > 0 {
		fmt.Fprintf(&b, "# Line %d: %s\n", lineNum, line.Content)
	}

	b.WriteString("# Lines above the blank line will be removed.\n")
	b.WriteString("#\n")
	b.WriteString("\n") // blank line separates template header from body

	if suggestion {
		b.WriteString("```suggestion\n")
		b.WriteString(line.Content + "\n")
		b.WriteString("```\n")
	}

	return b.String()
}

// ParseTemplate returns the body after the blank-line separator that follows
// the '#'-prefixed template header. Preserves '#' lines the user writes in
// their comment body. Falls back to the full content if no separator is found.
func ParseTemplate(content string) string {
	header, body, found := strings.Cut(content, "\n\n")
	if found {
		allHash := true
		for line := range strings.SplitSeq(header, "\n") {
			if !strings.HasPrefix(line, "#") {
				allHash = false
				break
			}
		}
		if allHash {
			return strings.TrimSpace(body)
		}
	}
	return strings.TrimSpace(content)
}

// Open spawns $EDITOR with a comment template and returns DraftAddedMsg on save.
// Returns nil if the user saves an empty body.
func Open(filename string, line diff.Line, suggestion bool) tea.Cmd {
	tmp, err := os.CreateTemp("", "pulley-comment-*.md")
	if err != nil {
		slog.Error("comment: create temp file", "err", err)
		return nil
	}

	_, err = tmp.WriteString(BuildTemplate(filename, line, suggestion))
	if closeErr := tmp.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		slog.Error("comment: write template", "err", err)
		_ = os.Remove(tmp.Name())
		return nil
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	tmpName := tmp.Name()
	return tea.ExecProcess(exec.Command(editor, tmpName), func(err error) tea.Msg { //nolint:gosec
		defer func() { _ = os.Remove(tmpName) }()
		if err != nil {
			slog.Error("comment: editor exited with error", "err", err)
			return nil
		}
		content, err := os.ReadFile(tmpName) //nolint:gosec
		if err != nil {
			slog.Error("comment: read temp file", "err", err)
			return nil
		}
		body := ParseTemplate(string(content))
		if body == "" {
			return nil
		}
		return DraftAddedMsg{
			Draft: github.DraftComment{
				Path:     filename,
				Position: line.DiffPosition,
				Body:     body,
			},
		}
	})
}
