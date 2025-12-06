# gitpulse

A terminal UI for monitoring and syncing multiple git repositories.

## Features

- Monitor status of multiple repos at a glance
- Fetch, sync (pull --rebase), and push with single keystrokes
- Smart upstream setup when tracking branch is missing
- Group repos by status (errors, behind, ahead, synced)
- 8 built-in color themes

## Installation

```bash
go install github.com/d12frosted/gitpulse@latest
```

Or build from source:

```bash
git clone https://github.com/d12frosted/gitpulse
cd gitpulse
go build
```

## Configuration

Config file location: `~/.config/gitpulse/config.toml`

```toml
# Color theme: dracula, nord, catppuccin, gruvbox, tokyonight, mono, jrpg-dark, jrpg-light
theme = "dracula"

# Repository paths to monitor
repos = [
    "~/Developer/project1",
    "~/Developer/project2",
    "~/work/important-repo",
]
```

Run `gitpulse --init` to generate an example config.

## Keybindings

| Key | Action |
|-----|--------|
| `j` / `k` | Move cursor down / up |
| `f` | Fetch selected repo |
| `F` | Fetch all repos |
| `s` | Sync selected repo (fetch + pull --rebase) |
| `S` | Sync all repos |
| `p` | Push selected repo |
| `P` | Push all repos |
| `u` | Set upstream branch |
| `r` | Refresh all statuses |
| `g` | Toggle grouping by status |
| `q` | Quit |

### Smart upstream setup

When you press `f`, `s`, or `u` on a repo without a tracking branch:

1. If remotes exist: shows a modal to select which remote branch to track
2. If no remotes: prompts to add an origin remote URL
3. After setup, continues with the original action (fetch/sync)

## Themes

Available themes:
- `dracula` (default)
- `nord`
- `catppuccin`
- `gruvbox`
- `tokyonight`
- `mono`
- `jrpg-dark`
- `jrpg-light`

## Status indicators

| Indicator | Meaning |
|-----------|---------|
| `*` | Uncommitted changes |
| `↑N` | N commits ahead of upstream |
| `↓N` | N commits behind upstream |
| `✓ synced` | Up to date with upstream |
| `○ no remote` | No upstream configured |
| `✗ error` | Error accessing repo |
