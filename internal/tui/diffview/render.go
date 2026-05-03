package diffview

import (
	"fmt"
	"strings"

	"github.com/rcorre/pulley/internal/diff"
	"github.com/rcorre/pulley/internal/syntax"
)

// Render converts a FileDiff into styled display lines and hunk row indices.
// hunkRows contains the line index of each hunk header within the returned slice.
func Render(f diff.FileDiff, cfg Config, hlr *syntax.Highlighter) (lines []string, hunkRows []int) {
	filename := f.NewName
	if filename == "" {
		filename = f.OldName
	}

	numWidth := lineNumWidth(f)

	for _, hunk := range f.Hunks {
		hunkRows = append(hunkRows, len(lines))

		// Pad header to align with content: 2*numWidth (line nums) + 1 (space) + 1 (marker) + 1 (space after marker) + 1
		pad := strings.Repeat(" ", 2*numWidth+4)
		lines = append(lines, cfg.HunkFg.Render(pad+hunk.Header))

		contents := make([]string, len(hunk.Lines))
		for i, l := range hunk.Lines {
			contents[i] = l.Content
		}
		highlighted := hlr.HighlightLines(filename, contents)

		for i, l := range hunk.Lines {
			lines = append(lines, renderLine(l, highlighted[i], numWidth, cfg))
		}
	}

	return lines, hunkRows
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
