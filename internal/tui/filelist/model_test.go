package filelist_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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

func sendKey(m filelist.Model, k string) (filelist.Model, tea.Msg) {
	return sendKeyMsg(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
}

func sendKeyMsg(m filelist.Model, msg tea.KeyMsg) (filelist.Model, tea.Msg) {
	updated, cmd := m.Update(msg)
	var result tea.Msg
	if cmd != nil {
		result = cmd()
	}
	return updated, result
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

func TestNavDown_MovesAndEmitsMsg(t *testing.T) {
	m := newModel()
	assert.Equal(t, 0, m.SelectedIndex())

	m, msg := sendKey(m, "j")
	assert.Equal(t, 1, m.SelectedIndex())
	sel, ok := msg.(filelist.FileSelectedMsg)
	require.True(t, ok)
	assert.Equal(t, 1, sel.Index)
	assert.Equal(t, "modified.go", sel.File.NewName)
}

func TestNavUp_MovesAndEmitsMsg(t *testing.T) {
	m := newModel()
	m, _ = sendKey(m, "j")
	m, _ = sendKey(m, "j")
	assert.Equal(t, 2, m.SelectedIndex())

	m, msg := sendKey(m, "k")
	assert.Equal(t, 1, m.SelectedIndex())
	sel, ok := msg.(filelist.FileSelectedMsg)
	require.True(t, ok)
	assert.Equal(t, 1, sel.Index)
}

func TestNavDown_StopsAtEnd(t *testing.T) {
	m := newModel()
	m, _ = sendKey(m, "j")
	m, _ = sendKey(m, "j")
	assert.Equal(t, 2, m.SelectedIndex())

	m, msg := sendKey(m, "j")
	assert.Equal(t, 2, m.SelectedIndex())
	assert.Nil(t, msg)
}

func TestNavUp_StopsAtStart(t *testing.T) {
	m := newModel()
	assert.Equal(t, 0, m.SelectedIndex())

	m, msg := sendKey(m, "k")
	assert.Equal(t, 0, m.SelectedIndex())
	assert.Nil(t, msg)
}

func TestNavArrowKeys(t *testing.T) {
	m := newModel()
	m, msg := sendKeyMsg(m, tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, m.SelectedIndex())
	assert.IsType(t, filelist.FileSelectedMsg{}, msg)

	m, msg = sendKeyMsg(m, tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 0, m.SelectedIndex())
	assert.IsType(t, filelist.FileSelectedMsg{}, msg)
}

func TestSetFiles_ResetsCursor(t *testing.T) {
	m := newModel()
	m, _ = sendKey(m, "j")
	assert.Equal(t, 1, m.SelectedIndex())

	m.SetFiles(testFiles)
	assert.Equal(t, 0, m.SelectedIndex())
}

func TestEmptyFiles_NoMsg(t *testing.T) {
	m := filelist.New(config.Default())
	_, msg := sendKey(m, "j")
	assert.Nil(t, msg)
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
