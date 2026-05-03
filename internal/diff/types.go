package diff

// LineKind identifies the role of a diff line.
type LineKind int

// Line kind constants.
const (
	LineContext   LineKind = iota // unchanged context line (space prefix)
	LineAdd                       // added line (+ prefix)
	LineRemove                    // removed line (- prefix)
	LineNoNewline                 // "\ No newline at end of file"
)

// Line is a single line within a hunk.
type Line struct {
	Kind         LineKind
	Content      string // line content without the +/-/space prefix
	OldLine      int    // 1-based line number in old file; 0 if not applicable
	NewLine      int    // 1-based line number in new file; 0 if not applicable
	DiffPosition int    // 1-based GitHub diff position within this file
}

// Hunk is a contiguous block of changes within a file diff.
type Hunk struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Header   string // full @@ ... @@ header line
	Lines    []Line
}

// FileDiff holds the parsed diff for one file.
type FileDiff struct {
	OldName  string // empty for new files
	NewName  string // empty for deleted files
	IsNew    bool
	IsDelete bool
	IsRename bool
	IsBinary bool
	Hunks    []Hunk
}

// Name returns the display path for the file: NewName if set, otherwise OldName.
func (f FileDiff) Name() string {
	if f.NewName != "" {
		return f.NewName
	}
	return f.OldName
}
