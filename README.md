# pulley

Keyboard-driven terminal UI for reviewing GitHub pull requests. Stay in the
terminal while viewing syntax-highlighted diffs, reading existing review
comments inline, drafting new comments, and submitting reviews.

## Installation

```sh
go install github.com/rcorre/pulley/cmd/pulley@latest
```

Requires [`gh`](https://cli.github.com/) to be installed and authenticated
(`gh auth login`).

## Usage

```
pulley [<number> | <url> | <branch>]
```

With no arguments, opens the PR for the current branch. Accepts a PR number,
URL, or branch name - same as `gh pr view`.

### Key bindings

| Key | Action |
|-----|--------|
| `ctrl+n` | Next file |
| `ctrl+p` | Previous file |
| `j` / `down` | Move cursor down |
| `k` / `up` | Move cursor up |
| `ctrl+d` / `pgdown` | Page down |
| `ctrl+u` / `pgup` | Page up |
| `]` | Jump to next hunk |
| `[` | Jump to previous hunk |
| `c` | Draft a comment on the cursor line |
| `s` | Draft a suggestion on the cursor line |
| `S` | Open the review submission dialog |
| `q` / `ctrl+c` | Quit |
| `r` | Retry after a load error |

In the review dialog:

| Key | Action |
|-----|--------|
| `j` / `down` | Move to next action |
| `k` / `up` | Move to previous action |
| `enter` / `y` | Confirm selection (opens `$EDITOR` for the review body) |
| `esc` | Cancel |

## Configuration

Configuration is loaded from `~/.config/pulley/config.toml`. All fields are
optional; defaults use ANSI indexed colors 0-15 for base16 palette
compatibility.

```toml
[colors]
add_fg     = 2   # ANSI index (integer) or "#rrggbb" (string)
add_bg     = 0
remove_fg  = 1
remove_bg  = 0
hunk_fg    = 4
cursor_bg  = 8
comment_fg = 3
comment_bg = 0
draft_fg   = 11
status_fg  = 0
status_bg  = 4

[keys]
quit          = ["q", "ctrl+c"]
up            = ["up", "k"]
down          = ["down", "j"]
page_up       = ["pgup", "ctrl+u"]
page_down     = ["pgdown", "ctrl+d"]
next_file     = ["ctrl+n"]
prev_file     = ["ctrl+p"]
comment       = ["c"]
suggestion    = ["s"]
submit_review = ["S"]
next_hunk     = ["]"]
prev_hunk     = ["["]
confirm       = ["enter", "y"]
cancel        = ["esc"]
retry         = ["r"]
```

## Shell completion

```sh
# bash
source <(pulley completion bash)

# zsh
source <(pulley completion zsh)

# fish
pulley completion fish | source
```

## Environment variables

| Variable | Description |
|----------|-------------|
| `PULLEY_LOG` | Path to a log file. If unset, logging is disabled. |
| `EDITOR` | Editor used for drafting comments and the review body. Defaults to `vi`. |
