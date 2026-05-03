package diff

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// join builds a diff string from lines, each terminated with a newline.
func join(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}

func TestParseAddRemove(t *testing.T) {
	raw := join(
		"diff --git a/foo.go b/foo.go",
		"index abc1234..def5678 100644",
		"--- a/foo.go",
		"+++ b/foo.go",
		"@@ -1,3 +1,3 @@",
		" package main",
		"-func Old() {}",
		"+func New() {}",
	)

	files, err := Parse(raw)
	require.NoError(t, err)
	require.Len(t, files, 1)

	f := files[0]
	assert.Equal(t, "foo.go", f.OldName)
	assert.Equal(t, "foo.go", f.NewName)
	assert.False(t, f.IsNew)
	assert.False(t, f.IsDelete)
	assert.False(t, f.IsRename)
	assert.False(t, f.IsBinary)
	require.Len(t, f.Hunks, 1)

	h := f.Hunks[0]
	assert.Equal(t, 1, h.OldStart)
	assert.Equal(t, 3, h.OldCount)
	assert.Equal(t, 1, h.NewStart)
	assert.Equal(t, 3, h.NewCount)
	require.Len(t, h.Lines, 3)

	ctx := h.Lines[0]
	assert.Equal(t, LineContext, ctx.Kind)
	assert.Equal(t, "package main", ctx.Content)
	assert.Equal(t, 1, ctx.OldLine)
	assert.Equal(t, 1, ctx.NewLine)

	rem := h.Lines[1]
	assert.Equal(t, LineRemove, rem.Kind)
	assert.Equal(t, "func Old() {}", rem.Content)
	assert.Equal(t, 2, rem.OldLine)
	assert.Equal(t, 0, rem.NewLine)

	add := h.Lines[2]
	assert.Equal(t, LineAdd, add.Kind)
	assert.Equal(t, "func New() {}", add.Content)
	assert.Equal(t, 0, add.OldLine)
	assert.Equal(t, 2, add.NewLine)
}

func TestParseNewFile(t *testing.T) {
	raw := join(
		"diff --git a/new.go b/new.go",
		"new file mode 100644",
		"index 0000000..abc1234",
		"--- /dev/null",
		"+++ b/new.go",
		"@@ -0,0 +1,2 @@",
		"+line1",
		"+line2",
	)

	files, err := Parse(raw)
	require.NoError(t, err)
	require.Len(t, files, 1)

	f := files[0]
	assert.Empty(t, f.OldName)
	assert.Equal(t, "new.go", f.NewName)
	assert.True(t, f.IsNew)
	assert.False(t, f.IsDelete)
	require.Len(t, f.Hunks, 1)

	h := f.Hunks[0]
	assert.Equal(t, 0, h.OldStart)
	assert.Equal(t, 0, h.OldCount)
	assert.Equal(t, 1, h.NewStart)
	assert.Equal(t, 2, h.NewCount)
	require.Len(t, h.Lines, 2)

	assert.Equal(t, LineAdd, h.Lines[0].Kind)
	assert.Equal(t, "line1", h.Lines[0].Content)
	assert.Equal(t, 1, h.Lines[0].NewLine)

	assert.Equal(t, LineAdd, h.Lines[1].Kind)
	assert.Equal(t, "line2", h.Lines[1].Content)
	assert.Equal(t, 2, h.Lines[1].NewLine)
}

func TestParseDeleteFile(t *testing.T) {
	raw := join(
		"diff --git a/old.go b/old.go",
		"deleted file mode 100644",
		"index abc1234..0000000",
		"--- a/old.go",
		"+++ /dev/null",
		"@@ -1,2 +0,0 @@",
		"-line1",
		"-line2",
	)

	files, err := Parse(raw)
	require.NoError(t, err)
	require.Len(t, files, 1)

	f := files[0]
	assert.Equal(t, "old.go", f.OldName)
	assert.Empty(t, f.NewName)
	assert.False(t, f.IsNew)
	assert.True(t, f.IsDelete)
	require.Len(t, f.Hunks, 1)

	h := f.Hunks[0]
	assert.Equal(t, 1, h.OldStart)
	assert.Equal(t, 2, h.OldCount)
	assert.Equal(t, 0, h.NewStart)
	assert.Equal(t, 0, h.NewCount)
	require.Len(t, h.Lines, 2)

	assert.Equal(t, LineRemove, h.Lines[0].Kind)
	assert.Equal(t, "line1", h.Lines[0].Content)
	assert.Equal(t, 1, h.Lines[0].OldLine)

	assert.Equal(t, LineRemove, h.Lines[1].Kind)
	assert.Equal(t, "line2", h.Lines[1].Content)
	assert.Equal(t, 2, h.Lines[1].OldLine)
}

func TestParseRename(t *testing.T) {
	// Pure rename (100% similarity, no content change - no hunks).
	raw := join(
		"diff --git a/old.go b/new.go",
		"similarity index 100%",
		"rename from old.go",
		"rename to new.go",
	)

	files, err := Parse(raw)
	require.NoError(t, err)
	require.Len(t, files, 1)

	f := files[0]
	assert.Equal(t, "old.go", f.OldName)
	assert.Equal(t, "new.go", f.NewName)
	assert.True(t, f.IsRename)
	assert.False(t, f.IsNew)
	assert.False(t, f.IsDelete)
	assert.Empty(t, f.Hunks)
}

func TestParseRenameWithChanges(t *testing.T) {
	raw := join(
		"diff --git a/old.go b/new.go",
		"similarity index 80%",
		"rename from old.go",
		"rename to new.go",
		"index abc..def 100644",
		"--- a/old.go",
		"+++ b/new.go",
		"@@ -1,2 +1,2 @@",
		" same",
		"-old content",
		"+new content",
	)

	files, err := Parse(raw)
	require.NoError(t, err)
	require.Len(t, files, 1)

	f := files[0]
	assert.Equal(t, "old.go", f.OldName)
	assert.Equal(t, "new.go", f.NewName)
	assert.True(t, f.IsRename)
	require.Len(t, f.Hunks, 1)
	assert.Len(t, f.Hunks[0].Lines, 3)
}

func TestParseMultiHunk(t *testing.T) {
	raw := join(
		"diff --git a/foo.go b/foo.go",
		"index abc..def 100644",
		"--- a/foo.go",
		"+++ b/foo.go",
		"@@ -1,3 +1,3 @@",
		" ctx1",
		"-old1",
		"+new1",
		"@@ -10,3 +10,3 @@",
		" ctx2",
		"-old2",
		"+new2",
	)

	files, err := Parse(raw)
	require.NoError(t, err)
	require.Len(t, files, 1)

	f := files[0]
	require.Len(t, f.Hunks, 2)

	h0 := f.Hunks[0]
	assert.Equal(t, 1, h0.OldStart)
	assert.Equal(t, 3, h0.OldCount)
	require.Len(t, h0.Lines, 3)
	assert.Equal(t, LineContext, h0.Lines[0].Kind)
	assert.Equal(t, LineRemove, h0.Lines[1].Kind)
	assert.Equal(t, LineAdd, h0.Lines[2].Kind)

	h1 := f.Hunks[1]
	assert.Equal(t, 10, h1.OldStart)
	require.Len(t, h1.Lines, 3)
	assert.Equal(t, LineContext, h1.Lines[0].Kind)
	assert.Equal(t, LineRemove, h1.Lines[1].Kind)
	assert.Equal(t, LineAdd, h1.Lines[2].Kind)

	// Line numbers in second hunk start from OldStart/NewStart.
	assert.Equal(t, 10, h1.Lines[0].OldLine)
	assert.Equal(t, 10, h1.Lines[0].NewLine)
	assert.Equal(t, 11, h1.Lines[1].OldLine)
	assert.Equal(t, 11, h1.Lines[2].NewLine)
}

func TestParseBinary(t *testing.T) {
	raw := join(
		"diff --git a/image.png b/image.png",
		"index abc1234..def5678 100644",
		"Binary files a/image.png and b/image.png differ",
	)

	files, err := Parse(raw)
	require.NoError(t, err)
	require.Len(t, files, 1)

	f := files[0]
	assert.Equal(t, "image.png", f.OldName)
	assert.Equal(t, "image.png", f.NewName)
	assert.True(t, f.IsBinary)
	assert.Empty(t, f.Hunks)
}

func TestParseNoNewline(t *testing.T) {
	raw := join(
		"diff --git a/foo.go b/foo.go",
		"index abc..def 100644",
		"--- a/foo.go",
		"+++ b/foo.go",
		"@@ -1,2 +1,2 @@",
		" context",
		"-old",
		`\ No newline at end of file`,
		"+new",
		`\ No newline at end of file`,
	)

	files, err := Parse(raw)
	require.NoError(t, err)
	require.Len(t, files, 1)

	h := files[0].Hunks[0]
	require.Len(t, h.Lines, 5)
	assert.Equal(t, LineContext, h.Lines[0].Kind)
	assert.Equal(t, LineRemove, h.Lines[1].Kind)
	assert.Equal(t, LineNoNewline, h.Lines[2].Kind)
	assert.Equal(t, LineAdd, h.Lines[3].Kind)
	assert.Equal(t, LineNoNewline, h.Lines[4].Kind)
}

func TestParseMultipleFiles(t *testing.T) {
	raw := join(
		"diff --git a/a.go b/a.go",
		"index abc..def 100644",
		"--- a/a.go",
		"+++ b/a.go",
		"@@ -1 +1 @@",
		"-old a",
		"+new a",
		"diff --git a/b.go b/b.go",
		"index abc..def 100644",
		"--- a/b.go",
		"+++ b/b.go",
		"@@ -1 +1 @@",
		"-old b",
		"+new b",
	)

	files, err := Parse(raw)
	require.NoError(t, err)
	require.Len(t, files, 2)
	assert.Equal(t, "a.go", files[0].NewName)
	assert.Equal(t, "b.go", files[1].NewName)
}

func TestPositionMapping(t *testing.T) {
	// Diff with two hunks to verify positions span across hunks.
	// DiffPosition matches GitHub's position scheme: position 1 is the first
	// content line of the first hunk; hunk headers are not counted.
	//
	// Expected positions:
	//   pos 1: " ctx1"  old=1 new=1
	//   pos 2: "-old1"  old=2
	//   pos 3: "+new1"  new=2
	//   pos 4: " ctx2"  old=10 new=10
	//   pos 5: "-old2"  old=11
	//   pos 6: "+new2"  new=11
	raw := join(
		"diff --git a/foo.go b/foo.go",
		"index abc..def 100644",
		"--- a/foo.go",
		"+++ b/foo.go",
		"@@ -1,3 +1,3 @@",
		" ctx1",
		"-old1",
		"+new1",
		"@@ -10,3 +10,3 @@",
		" ctx2",
		"-old2",
		"+new2",
	)

	files, err := Parse(raw)
	require.NoError(t, err)
	require.Len(t, files, 1)
	fd := files[0]

	require.Len(t, fd.Hunks, 2)

	// Verify DiffPosition values are assigned correctly.
	h0 := fd.Hunks[0]
	assert.Equal(t, 1, h0.Lines[0].DiffPosition) // ctx1
	assert.Equal(t, 2, h0.Lines[1].DiffPosition) // old1
	assert.Equal(t, 3, h0.Lines[2].DiffPosition) // new1

	h1 := fd.Hunks[1]
	assert.Equal(t, 4, h1.Lines[0].DiffPosition) // ctx2
	assert.Equal(t, 5, h1.Lines[1].DiffPosition) // old2
	assert.Equal(t, 6, h1.Lines[2].DiffPosition) // new2

	// CommentPosition: new-file line -> diff position.
	pos, ok := fd.CommentPosition(1) // ctx1 is new-file line 1
	assert.True(t, ok)
	assert.Equal(t, 1, pos)

	pos, ok = fd.CommentPosition(2) // new1 is new-file line 2
	assert.True(t, ok)
	assert.Equal(t, 3, pos)

	pos, ok = fd.CommentPosition(10) // ctx2 is new-file line 10
	assert.True(t, ok)
	assert.Equal(t, 4, pos)

	pos, ok = fd.CommentPosition(11) // new2 is new-file line 11
	assert.True(t, ok)
	assert.Equal(t, 6, pos)

	_, ok = fd.CommentPosition(99)
	assert.False(t, ok)

	// LineForPosition: diff position -> Line.
	line, ok := fd.LineForPosition(1)
	assert.True(t, ok)
	assert.Equal(t, LineContext, line.Kind)
	assert.Equal(t, "ctx1", line.Content)

	line, ok = fd.LineForPosition(3)
	assert.True(t, ok)
	assert.Equal(t, LineAdd, line.Kind)
	assert.Equal(t, "new1", line.Content)

	line, ok = fd.LineForPosition(5)
	assert.True(t, ok)
	assert.Equal(t, LineRemove, line.Kind)
	assert.Equal(t, "old2", line.Content)

	_, ok = fd.LineForPosition(99)
	assert.False(t, ok)

	// CommentPosition and LineForPosition are inverses for add/context lines.
	for _, newL := range []int{1, 2, 10, 11} {
		p, found := fd.CommentPosition(newL)
		require.True(t, found, "CommentPosition(%d) not found", newL)
		l, found := fd.LineForPosition(p)
		require.True(t, found, "LineForPosition(%d) not found", p)
		assert.Equal(t, newL, l.NewLine)
	}
}
