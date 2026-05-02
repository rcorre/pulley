package diff

// CommentPosition returns the GitHub diff position for the given new-file
// line number. It matches LineAdd and LineContext lines only (lines that exist
// in the new file). Returns 0, false if the line number is not found.
func (fd FileDiff) CommentPosition(newLine int) (int, bool) {
	for _, h := range fd.Hunks {
		for _, l := range h.Lines {
			if l.NewLine == newLine && (l.Kind == LineAdd || l.Kind == LineContext) {
				return l.DiffPosition, true
			}
		}
	}
	return 0, false
}

// LineForPosition returns the Line at the given 1-based diff position.
// Returns the zero Line and false if not found.
func (fd FileDiff) LineForPosition(pos int) (Line, bool) {
	for _, h := range fd.Hunks {
		for _, l := range h.Lines {
			if l.DiffPosition == pos {
				return l, true
			}
		}
	}
	return Line{}, false
}
