package filelist_test

import (
	"strings"
	"testing"

	"github.com/rcorre/pulley/internal/config"
	"github.com/rcorre/pulley/internal/diff"
	"github.com/rcorre/pulley/internal/tui/filelist"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testFiles = []diff.FileDiff{
	{NewName: "added.go", IsNew: true},
	{OldName: "modified.go", NewName: "modified.go"},
	{OldName: "deleted.go", IsDelete: true},
}

func newModel() filelist.Model {
	m := filelist.New(config.Default())
	m.SetFiles(testFiles)
	return m
}

func TestView_ShowsFileNames(t *testing.T) {
	m := newModel()
	view := m.View()
	assert.Contains(t, view, "added.go")
	assert.Contains(t, view, "modified.go")
	assert.Contains(t, view, "deleted.go")
}

func TestView_ShowsStatusIndicators(t *testing.T) {
	m := newModel()
	lines := strings.Split(m.View(), "\n")
	require.Len(t, lines, 3)
	assert.Contains(t, lines[0], "A")
	assert.Contains(t, lines[1], "M")
	assert.Contains(t, lines[2], "D")
}

func TestNextFile_MovesAndEmitsMsg(t *testing.T) {
	m := newModel()
	assert.Equal(t, 0, m.SelectedIndex())

	cmd := m.NextFile()
	assert.Equal(t, 1, m.SelectedIndex())
	require.NotNil(t, cmd)
	sel, ok := cmd().(filelist.FileSelectedMsg)
	require.True(t, ok)
	assert.Equal(t, 1, sel.Index)
	assert.Equal(t, "modified.go", sel.File.NewName)
}

func TestPrevFile_MovesAndEmitsMsg(t *testing.T) {
	m := newModel()
	m.NextFile()
	m.NextFile()
	assert.Equal(t, 2, m.SelectedIndex())

	cmd := m.PrevFile()
	assert.Equal(t, 1, m.SelectedIndex())
	require.NotNil(t, cmd)
	sel, ok := cmd().(filelist.FileSelectedMsg)
	require.True(t, ok)
	assert.Equal(t, 1, sel.Index)
}

func TestNextFile_WrapsAtEnd(t *testing.T) {
	m := newModel()
	m.NextFile()
	m.NextFile()
	assert.Equal(t, 2, m.SelectedIndex())

	cmd := m.NextFile()
	assert.Equal(t, 0, m.SelectedIndex())
	require.NotNil(t, cmd)
	sel, ok := cmd().(filelist.FileSelectedMsg)
	require.True(t, ok)
	assert.Equal(t, 0, sel.Index)
}

func TestPrevFile_WrapsAtStart(t *testing.T) {
	m := newModel()
	assert.Equal(t, 0, m.SelectedIndex())

	cmd := m.PrevFile()
	assert.Equal(t, 2, m.SelectedIndex())
	require.NotNil(t, cmd)
	sel, ok := cmd().(filelist.FileSelectedMsg)
	require.True(t, ok)
	assert.Equal(t, 2, sel.Index)
}

func TestNextFile_EmptyFiles_ReturnsNil(t *testing.T) {
	m := filelist.New(config.Default())
	cmd := m.NextFile()
	assert.Nil(t, cmd)
}

func TestPrevFile_EmptyFiles_ReturnsNil(t *testing.T) {
	m := filelist.New(config.Default())
	cmd := m.PrevFile()
	assert.Nil(t, cmd)
}

func TestSetFiles_ResetsCursor(t *testing.T) {
	m := newModel()
	m.NextFile()
	assert.Equal(t, 1, m.SelectedIndex())

	m.SetFiles(testFiles)
	assert.Equal(t, 0, m.SelectedIndex())
}

func TestDisplayName_UsesNewName(t *testing.T) {
	m := filelist.New(config.Default())
	m.SetFiles([]diff.FileDiff{
		{OldName: "old.go", NewName: "new.go", IsRename: true},
	})
	assert.Contains(t, m.View(), "new.go")
}

func TestDisplayName_FallsBackToOldName(t *testing.T) {
	m := filelist.New(config.Default())
	m.SetFiles([]diff.FileDiff{
		{OldName: "deleted.go", IsDelete: true},
	})
	assert.Contains(t, m.View(), "deleted.go")
}
