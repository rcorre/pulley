package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefault_Sane(t *testing.T) {
	cfg := Default()

	assert.NotEmpty(t, cfg.Colors.AddFg.String(), "add_fg should have a default")
	assert.NotEmpty(t, cfg.Colors.RemoveFg.String(), "remove_fg should have a default")
	assert.NotEmpty(t, cfg.Colors.HunkFg.String(), "hunk_fg should have a default")
	assert.NotEmpty(t, cfg.Colors.StatusFg.String(), "status_fg should have a default")
	assert.NotEmpty(t, cfg.Colors.StatusBg.String(), "status_bg should have a default")

	assert.NotEmpty(t, cfg.Keys.Quit, "quit keys should have defaults")
	assert.NotEmpty(t, cfg.Keys.Up, "up keys should have defaults")
	assert.NotEmpty(t, cfg.Keys.Down, "down keys should have defaults")
	assert.NotEmpty(t, cfg.Keys.Tab, "tab keys should have defaults")
	assert.NotEmpty(t, cfg.Keys.Comment, "comment keys should have defaults")
	assert.NotEmpty(t, cfg.Keys.SubmitReview, "submit_review keys should have defaults")
}

func TestLoad_MissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.toml")
	require.NoError(t, err)
	assert.Equal(t, Default(), cfg)
}

func TestLoad_PartialMerge(t *testing.T) {
	content := `
[colors]
add_fg = 10

[keys]
quit = ["ctrl+c"]
`
	path := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	cfg, err := Load(path)
	require.NoError(t, err)

	// Overridden values
	assert.Equal(t, "10", cfg.Colors.AddFg.String())
	assert.Equal(t, []string{"ctrl+c"}, cfg.Keys.Quit)

	// Defaults preserved for fields not in the file
	assert.Equal(t, Default().Colors.RemoveFg.String(), cfg.Colors.RemoveFg.String())
	assert.Equal(t, Default().Keys.Up, cfg.Keys.Up)
}

func TestColorValue_ParseInt(t *testing.T) {
	content := `
[colors]
add_fg = 2
remove_fg = 255
`
	path := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	cfg, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, "2", cfg.Colors.AddFg.String())
	assert.Equal(t, "255", cfg.Colors.RemoveFg.String())
}

func TestColorValue_ParseHex(t *testing.T) {
	content := `
[colors]
add_fg = "#50fa7b"
`
	path := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	cfg, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, "#50fa7b", cfg.Colors.AddFg.String())
}

func TestColorValue_InvalidIndex(t *testing.T) {
	content := `
[colors]
add_fg = 300
`
	path := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	_, err := Load(path)
	assert.Error(t, err)
}
