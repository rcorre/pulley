package diffview

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/rcorre/pulley/internal/diff"
	"github.com/rcorre/pulley/internal/syntax"
)

// Comment is an inline review comment to display below a diff line.
type Comment struct {
	Author   string
	Body     string
	Position int // GitHub diff position
}

// Render converts a FileDiff into styled display lines and hunk row indices.
// hunkRows contains the line index of each hunk header within the returned slice.
// comments are rendered inline below their target diff lines.
func Render(f diff.FileDiff, cfg Config, hlr *syntax.Highlighter, comments []Comment) (lines []string, hunkRows []int) {
	byPos := make(map[int][]Comment, len(comments))
	for _, c := range comments {
		byPos[c.Position] = append(byPos[c.Position], c)
	}

	filename := f.Name()
	numWidth := lineNumWidth(f)
	// Pad aligns with content: 2*numWidth (line nums) + 1 (space) + 1 (marker) + 1 (space after marker) + 1
	indent := strings.Repeat(" ", 2*numWidth+4)
	commentStyle := cfg.CommentFg.Inherit(cfg.CommentBg)

	for _, hunk := range f.Hunks {
		hunkRows = append(hunkRows, len(lines))
		lines = append(lines, cfg.HunkFg.Render(indent+hunk.Header))

		contents := make([]string, len(hunk.Lines))
		for i, l := range hunk.Lines {
			contents[i] = l.Content
		}
		highlighted := hlr.HighlightLines(filename, contents)

		for i, l := range hunk.Lines {
			lines = append(lines, renderLine(l, highlighted[i], numWidth, cfg))
			if l.DiffPosition > 0 {
				for _, c := range byPos[l.DiffPosition] {
					lines = append(lines, renderComment(c, indent, commentStyle)...)
				}
			}
		}
	}

	return lines, hunkRows
}

func renderComment(c Comment, indent string, style lipgloss.Style) []string {
	var result []string
	result = append(result, style.Render(indent+"@ "+c.Author))
	for line := range strings.SplitSeq(c.Body, "\n") {
		result = append(result, style.Render(indent+"  "+line))
	}
	return result
}

func renderLine(l diff.Line, highlighted string, numWidth int, cfg Config) string {
	fmtNum := func(n int) string {
		if n > 0 {
			return fmt.Sprintf("%*d", numWidth, n)
		}
		return strings.Repeat(" ", numWidth)
	}
	gutter := cfg.LineNum.Render(fmtNum(l.OldLine) + " " + fmtNum(l.NewLine))

	switch l.Kind {
	case diff.LineAdd:
		return gutter + " " + cfg.AddFg.Render("+") + " " + cfg.AddBg.Render(highlighted)
	case diff.LineRemove:
		return gutter + " " + cfg.RemoveFg.Render("-") + " " + cfg.RemoveBg.Render(highlighted)
	default:
		return gutter + "   " + highlighted
	}
}

// lineNumWidth computes the minimum column width needed to display all line numbers.
func lineNumWidth(f diff.FileDiff) int {
	maxLine := 1
	for _, h := range f.Hunks {
		for _, l := range h.Lines {
			if l.OldLine > maxLine {
				maxLine = l.OldLine
			}
			if l.NewLine > maxLine {
				maxLine = l.NewLine
			}
		}
	}
	return len(fmt.Sprintf("%d", maxLine))
}
