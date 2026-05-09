package config

import (
	"errors"
	"os"

	"github.com/BurntSushi/toml"
)

// Config holds the full application configuration.
type Config struct {
	Colors ColorConfig `toml:"colors"`
	Keys   KeyConfig   `toml:"keys"`
}

// ColorConfig holds color settings for UI elements.
// All defaults use ANSI indexed values 0-15 for base16 palette compatibility.
type ColorConfig struct {
	AddFg    ColorValue `toml:"add_fg"`
	RemoveFg ColorValue `toml:"remove_fg"`
	HunkFg     ColorValue `toml:"hunk_fg"`
	FileHeader ColorValue `toml:"file_header"`
	LineNum    ColorValue `toml:"line_num"`
	CursorBg   ColorValue `toml:"cursor_bg"`
	CommentFg  ColorValue `toml:"comment_fg"`
	CommentBg  ColorValue `toml:"comment_bg"`
	DraftFg    ColorValue `toml:"draft_fg"`
	FileModFg  ColorValue `toml:"file_mod_fg"`
	StatusFg   ColorValue `toml:"status_fg"`
	StatusBg   ColorValue `toml:"status_bg"`
}

// KeyConfig holds key bindings as string slices to support multiple keys per action.
type KeyConfig struct {
	Quit         []string `toml:"quit"`
	Up           []string `toml:"up"`
	Down         []string `toml:"down"`
	PageUp       []string `toml:"page_up"`
	PageDown     []string `toml:"page_down"`
	HalfPageUp   []string `toml:"half_page_up"`
	HalfPageDown []string `toml:"half_page_down"`
	Comment      []string `toml:"comment"`
	Suggestion   []string `toml:"suggestion"`
	SubmitReview []string `toml:"submit_review"`
	NextFile     []string `toml:"next_file"`
	PrevFile     []string `toml:"prev_file"`
	NextHunk     []string `toml:"next_hunk"`
	PrevHunk     []string `toml:"prev_hunk"`
	Confirm      []string `toml:"confirm"`
	Cancel       []string `toml:"cancel"`
	Retry        []string `toml:"retry"`
	Suspend      []string `toml:"suspend"`
}

// Default returns a Config populated with sensible defaults.
func Default() Config {
	return Config{
		Colors: defaultColors(),
		Keys:   defaultKeys(),
	}
}

// Load reads a TOML config file at path and merges it over the defaults.
// If path does not exist, the defaults are returned without error.
func Load(path string) (Config, error) {
	cfg := Default()
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	return cfg, err
}
