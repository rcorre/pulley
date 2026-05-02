// Package diff parses unified git diffs into structured data.
package diff

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var hunkRe = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)

// Parse parses a unified git diff string and returns one FileDiff per file.
// The DiffPosition field on each Line reflects GitHub's diff position scheme:
// position 1 is the first hunk header; each subsequent hunk header and diff
// line increments the counter, which resets for each new file.
func Parse(raw string) ([]FileDiff, error) {
	var files []FileDiff
	var cur *FileDiff
	curHunkIdx := -1
	diffPos := 0
	oldLine, newLine := 0, 0

	for line := range strings.SplitSeq(raw, "\n") {
		switch {
		case strings.HasPrefix(line, "diff --git "):
			if cur != nil {
				files = append(files, *cur)
			}
			cur = &FileDiff{}
			curHunkIdx = -1
			diffPos = 0
			// Extract names from "diff --git a/OLD b/NEW".
			// This is an initial best-guess; --- and +++ lines take precedence.
			rest := strings.TrimPrefix(line, "diff --git ")
			if before, after, ok := strings.Cut(rest, " b/"); ok {
				cur.OldName = strings.TrimPrefix(before, "a/")
				cur.NewName = after
			}

		case cur == nil:
			// ignore lines before the first file header

		case strings.HasPrefix(line, "new file mode"):
			cur.IsNew = true

		case strings.HasPrefix(line, "deleted file mode"):
			cur.IsDelete = true

		case strings.HasPrefix(line, "rename from "):
			cur.IsRename = true
			cur.OldName = strings.TrimPrefix(line, "rename from ")

		case strings.HasPrefix(line, "rename to "):
			cur.NewName = strings.TrimPrefix(line, "rename to ")

		case strings.HasPrefix(line, "Binary files"):
			cur.IsBinary = true

		// Guard --- and +++ with curHunkIdx < 0 so hunk content lines
		// that start with --- or +++ are handled as diff lines, not headers.
		case strings.HasPrefix(line, "--- ") && curHunkIdx < 0:
			name := strings.TrimPrefix(line, "--- ")
			if name == "/dev/null" {
				cur.IsNew = true
				cur.OldName = ""
			} else {
				cur.OldName = strings.TrimPrefix(name, "a/")
			}

		case strings.HasPrefix(line, "+++ ") && curHunkIdx < 0:
			name := strings.TrimPrefix(line, "+++ ")
			if name == "/dev/null" {
				cur.IsDelete = true
				cur.NewName = ""
			} else {
				cur.NewName = strings.TrimPrefix(name, "b/")
			}

		case strings.HasPrefix(line, "@@ "):
			m := hunkRe.FindStringSubmatch(line)
			if m == nil {
				return nil, fmt.Errorf("malformed hunk header: %q", line)
			}
			diffPos++
			oldStart, _ := strconv.Atoi(m[1])
			oldCount := 1
			if m[2] != "" {
				oldCount, _ = strconv.Atoi(m[2])
			}
			newStart, _ := strconv.Atoi(m[3])
			newCount := 1
			if m[4] != "" {
				newCount, _ = strconv.Atoi(m[4])
			}
			cur.Hunks = append(cur.Hunks, Hunk{
				OldStart: oldStart,
				OldCount: oldCount,
				NewStart: newStart,
				NewCount: newCount,
				Header:   line,
			})
			curHunkIdx = len(cur.Hunks) - 1
			oldLine = oldStart
			newLine = newStart

		case curHunkIdx >= 0 && line != "":
			diffPos++
			h := &cur.Hunks[curHunkIdx]
			var l Line
			l.DiffPosition = diffPos
			switch line[0] {
			case '+':
				l = Line{Kind: LineAdd, Content: line[1:], NewLine: newLine, DiffPosition: diffPos}
				newLine++
			case '-':
				l = Line{Kind: LineRemove, Content: line[1:], OldLine: oldLine, DiffPosition: diffPos}
				oldLine++
			case '\\':
				l = Line{Kind: LineNoNewline, Content: line, DiffPosition: diffPos}
			default: // space prefix = context line
				l = Line{Kind: LineContext, Content: line[1:], OldLine: oldLine, NewLine: newLine, DiffPosition: diffPos}
				oldLine++
				newLine++
			}
			h.Lines = append(h.Lines, l)
		}
	}

	if cur != nil {
		files = append(files, *cur)
	}
	return files, nil
}
