# Pulley — Interactive PR Review TUI

## Context

Build a keyboard-centric terminal UI for reviewing GitHub pull requests. The motivation is to stay in the terminal during code review — viewing syntax-highlighted diffs, reading existing comments inline, drafting new comments/suggestions, and submitting reviews without leaving the CLI. Defaults to indexed terminal colors so the theme follows the user's base16 palette.

## Confirmed Requirements

- **Unified diff only** (no side-by-side)
- **File list panel** alongside the diff (persistent sidebar)
- **Fetch & display** existing GitHub review comments inline (one-way read, no reply from TUI in v1)
- **Draft line comments and suggestions** locally, submit as part of a review
- **Submit PR review**: approve / request changes / comment
- **Invocation**: `pulley [<number> | <url> | <branch>]` — same as `gh pr view` behavior. No args = PR for current branch.
- **Configurable** colors (TOML, `~/.config/pulley/config.toml`) and keybinds
- **Base16 defaults**: all default colors use ANSI indexed 0-15

## Technology Stack

| Layer | Library | Why |
|---|---|---|
| TUI | `charmbracelet/bubbletea` | Elm architecture, dominant Go TUI framework |
| Styling | `charmbracelet/lipgloss` v2 | Layout, auto color profile detection |
| Components | `charmbracelet/bubbles` | viewport, textarea, list |
| Syntax | `alecthomas/chroma` v2 | Pure Go, Terminal256 formatter, no CGo |
| GitHub API | `cli/go-gh` v2 | Auth via gh token store, REST + GraphQL clients |
| CLI | `spf13/cobra` | Shell completion, battle-tested |
| Config | `BurntSushi/toml` | TOML parsing |
| Testing | `stretchr/testify`, `teatest` | Assertions + TUI testing |
| Linting | `golangci-lint` | staticcheck, errcheck, gosec, revive, testifylint |

## Project Structure

```
pulley/
├── cmd/pulley/main.go              # Cobra root command, config load, resolve PR, launch TUI
├── internal/
│   ├── config/
│   │   ├── config.go               # Config struct, Load(), Default(), TOML parsing
│   │   ├── colors.go               # ColorValue (int for indexed, string for hex)
│   │   └── keys.go                 # Default key mappings, action names
│   ├── github/
│   │   ├── client.go               # PRClient interface + go-gh implementation
│   │   ├── types.go                # PR, ReviewComment, DraftComment, ReviewEvent
│   │   └── resolve.go              # Resolve PR from number/URL/branch/empty
│   ├── diff/
│   │   ├── parse.go                # Parse unified diff → []FileDiff
│   │   ├── types.go                # FileDiff, Hunk, Line, LineKind
│   │   └── position.go             # Map (file, line, side) ↔ GitHub diff position
│   ├── syntax/
│   │   └── highlight.go            # Chroma wrapper, base16-compatible indexed style
│   └── tui/
│       ├── app.go                  # Root model: layout, focus routing, data flow
│       ├── msg.go                  # Custom tea.Msg types
│       ├── keymap.go               # Keymap struct from config → key.Binding
│       ├── style.go                # lipgloss Styles from config colors
│       ├── filelist/model.go       # Left panel: file list with status indicators
│       ├── diffview/
│       │   ├── model.go            # Scrollable diff viewport with cursor line
│       │   └── render.go           # Render diff + syntax + inline comments
│       ├── comment/editor.go       # Spawn $EDITOR for comment/suggestion drafting
│       ├── review/model.go         # Review submission dialog
│       └── statusbar/model.go      # Bottom bar: PR info, draft count, key hints
├── .github/workflows/ci.yml        # CI: build, test (-race), lint on push to main + PRs
├── .golangci.yml
├── go.mod
└── go.sum
```

## Architecture

### Component Hierarchy & Focus Model

```
app (root)
├── filelist   (left, ~25% width)   ← Tab toggles ←→
├── diffview   (right, ~75% width)  ← Tab toggles ←→
├── comment    ($EDITOR suspend)    ← 'c'/'s' suspends TUI, opens $EDITOR
├── review     (overlay, modal)     ← 'S' opens, Esc closes
└── statusbar  (bottom, full width)
```

Focus enum: `FocusFileList | FocusDiff | FocusReview`. Keys route to focused component only. Comment editing suspends the TUI entirely via `tea.Exec()`.

### Data Flow

1. **Startup**: Cobra parses args → load config → `github.Resolve()` gets PRRef → launch `tea.Program`
2. **Init**: fires async command that fetches PR metadata (GraphQL), diff (REST), comments (GraphQL) concurrently
3. **PRLoadedMsg**: parse diff via `diff.Parse()`, populate file list, select first file
4. **File selection**: syntax-highlight file via chroma, filter comments for file, render diff string, set as viewport content
5. **Add comment**: capture cursor line position → open comment overlay → on save, append to `Model.drafts`, re-render diff
6. **Submit review**: open review dialog → select action → fires `client.SubmitReview()` with all drafts → clear drafts on success

### Key Design Decisions

- **Pre-render per file**: On file selection, render the entire diff (with syntax + comments) into one string. Files are typically small; avoids lazy-rendering complexity.
- **Inline comments**: Comments rendered as indented blocks in the diff string itself (not floating overlays). Drafts shown with `[draft]` prefix, visually distinct.
- **$EDITOR for comments**: Press 'c'/'s' on a diff line → write temp file with comment template (line context as header comments, suggestion fence pre-filled for 's') → `tea.Exec()` suspends TUI and spawns `$EDITOR` → on exit, read file, strip template lines, create draft → TUI resumes. Same pattern as `git commit` / `gh pr create`.
- **Cursor line**: diffview tracks a highlighted cursor row (separate from scroll offset) for targeting comments. Auto-scrolls to keep cursor visible.
- **ColorValue**: TOML unmarshals integers as ANSI indexed (0-255), strings as hex. Defaults use only 0-15 (base16 palette). Lipgloss auto-degrades.
- **Diff parser**: Hand-written (~200 LOC). Tracks `DiffPosition` per line (needed for GitHub comment API). No external diff library needed.
- **GitHub API**: REST for diff + review submission. GraphQL for PR metadata + comments (single round-trip for nested data).
- **Draft storage**: In-memory for v1. Quit confirmation prompt if unsaved drafts exist. Status bar shows draft count. On-disk persistence (`~/.cache/pulley/`) planned as a follow-up unit.

## Workflow Per Unit

After each unit is implemented and tests pass:

1. Run `/simplify` to review changed code for reuse, quality, and efficiency - fix any issues found
2. Run `/review` to get a second pass on correctness and style
3. Run `go test -race ./...`, `golangci-lint run`, and `golangci-lint fmt`
4. Commit the unit
