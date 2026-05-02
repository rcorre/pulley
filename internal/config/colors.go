// Package config provides configuration loading and defaults for pulley.
package config

import (
	"fmt"
	"strconv"
)

// ColorValue represents a terminal color as either an ANSI indexed value
// (0-255) or a hex string ("#RRGGBB"). The zero value means no color.
type ColorValue struct {
	value string
}

// ANSI returns a ColorValue for an ANSI indexed color (0-255).
func ANSI(n int) ColorValue {
	return ColorValue{value: strconv.Itoa(n)}
}

// Hex returns a ColorValue for a hex color string (e.g., "#ff0000").
func Hex(s string) ColorValue {
	return ColorValue{value: s}
}

// String returns the color as a string compatible with lipgloss.Color.
// An empty string means no color.
func (c ColorValue) String() string {
	return c.value
}

// UnmarshalTOML implements toml.Unmarshaler.
// Integers are treated as ANSI indexed colors (0-255); strings as hex colors.
func (c *ColorValue) UnmarshalTOML(v any) error {
	switch val := v.(type) {
	case int64:
		if val < 0 || val > 255 {
			return fmt.Errorf("ANSI color index must be 0-255, got %d", val)
		}
		c.value = strconv.FormatInt(val, 10)
	case string:
		c.value = val
	default:
		return fmt.Errorf("color must be int (ANSI index) or string (hex), got %T", v)
	}
	return nil
}

func defaultColors() ColorConfig {
	return ColorConfig{
		AddFg:      ANSI(2), // green
		RemoveFg:   ANSI(1), // red
		HunkFg:     ANSI(6), // cyan
		FileHeader: ANSI(4), // blue
		LineNum:    ANSI(8), // bright black (dark gray)
		CursorBg:   ANSI(8), // bright black (subtle highlight)
		CommentFg:  ANSI(3), // yellow
		DraftFg:    ANSI(5), // magenta
		StatusFg:   ANSI(0), // black
		StatusBg:   ANSI(4), // blue
	}
}
