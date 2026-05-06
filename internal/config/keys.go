package config

func defaultKeys() KeyConfig {
	return KeyConfig{
		Quit:         []string{"q", "ctrl+c"},
		Up:           []string{"up", "k"},
		Down:         []string{"down", "j"},
		PageUp:       []string{"pgup", "ctrl+u"},
		PageDown:     []string{"pgdown", "ctrl+d"},
		Comment:      []string{"c"},
		Suggestion:   []string{"s"},
		SubmitReview: []string{"S"},
		NextFile:     []string{"ctrl+n"},
		PrevFile:     []string{"ctrl+p"},
		NextHunk:     []string{"]"},
		PrevHunk:     []string{"["},
		Confirm:      []string{"enter", "y"},
		Cancel:       []string{"esc"},
		Retry:        []string{"r"},
		Suspend:      []string{"ctrl+z"},
	}
}
