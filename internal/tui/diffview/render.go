package diffview

import (
	"fmt"
	"log/slog"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/rcorre/pulley/internal/diff"
	"github.com/rcorre/pulley/internal/syntax"
)

// Comment is an inline review comment to display below a diff line.
type Comment struct {
	Author   string
	Body     string
	Position int  // GitHub diff position
	Draft    bool // true for locally drafted, not yet submitted comments
}

// Render converts a FileDiff into styled display lines, hunk row indices, and a
// line map. lineMap is parallel to lines: each entry holds the diff.Line for that
// rendered row, or a zero-value Line (DiffPosition==0) for hunk headers and
// comment text rows.
func Render(f diff.FileDiff, cfg Config, hlr *syntax.Highlighter, comments []Comment) (lines []string, hunkRows []int, lineMap []diff.Line) {
	byPos := make(map[int][]Comment, len(comments))
	for _, c := range comments {
		slog.Debug("render: comment", "position", c.Position, "author", c.Author)
		byPos[c.Position] = append(byPos[c.Position], c)
	}
	slog.Debug("render: byPos keys", "count", len(byPos))

	filename := f.Name()
	numWidth := lineNumWidth(f)
	// Pad aligns with content: 2*numWidth (line nums) + 1 (space) + 1 (marker) + 1 (space after marker) + 1
	indent := strings.Repeat(" ", 2*numWidth+4)
	commentStyle := cfg.CommentFg.Inherit(cfg.CommentBg)
	draftStyle := cfg.DraftFg.Inherit(cfg.CommentBg)

	for _, hunk := range f.Hunks {
		hunkRows = append(hunkRows, len(lines))
		lines = append(lines, cfg.HunkFg.Render(indent+hunk.Header))
		lineMap = append(lineMap, diff.Line{})

		contents := make([]string, len(hunk.Lines))
		for i, l := range hunk.Lines {
			contents[i] = l.Content
		}
		highlighted := hlr.HighlightLines(filename, contents)

		for i, l := range hunk.Lines {
			lines = append(lines, renderLine(l, highlighted[i], numWidth, cfg))
			lineMap = append(lineMap, l)
			if l.DiffPosition > 0 {
				cs := byPos[l.DiffPosition]
				slog.Debug("render: line", "diffpos", l.DiffPosition, "comments", len(cs))
				for _, c := range cs {
					style := commentStyle
					if c.Draft {
						style = draftStyle
					}
					for _, row := range renderComment(c, indent, style) {
						lines = append(lines, row)
						lineMap = append(lineMap, diff.Line{})
					}
				}
			}
		}
	}

	return lines, hunkRows, lineMap
}

func renderComment(c Comment, indent string, style lipgloss.Style) []string {
	var result []string
	if c.Draft {
		result = append(result, style.Render(indent+"[draft]"))
	} else {
		result = append(result, style.Render(indent+"@ "+c.Author))
	}
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
