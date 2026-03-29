# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

**Obia** — an interactive TUI for Obsidian vault task management. Parses `- [ ]` tasks from all `.md` files in an Obsidian vault and displays them in a tabbed, vim-navigable interface with CalDAV sync support.

## Build & Run

```bash
go build -o obia .          # Build binary
go run .                    # Run directly
go test ./...               # Run all tests
go test ./internal/vault/   # Run tests for a specific package
```

## Architecture

Go project using Bubble Tea (charmbracelet) for the TUI. All application code lives under `internal/`:

- **`internal/config/`** — TOML config loading from `~/.config/obia/config.toml`
- **`internal/vault/`** — Scans `.md` files, parses `- [ ]` checkbox tasks, writes status changes back to files. Skips `.obsidian`, `.trash`, `.opencode`, `templates`, `extras` directories.
- **`internal/task/`** — Task data model (status, description, due date, source file + line, CalDAV UID, tags, wikilinks)
- **`internal/caldav/`** — CalDAV HTTP client (basic auth, VTODO build/parse, push/pull). Ported from the TypeScript Obsidian plugin at `~/code/projects/caldev-sync-plugin/caldav-sync/src/caldav.ts`.
- **`internal/tui/`** — Bubble Tea models: root app with tab switching, task list, add-task form, status bar, lipgloss styles

## Key Dependencies

- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/bubbles` — List, text input, viewport components
- `github.com/charmbracelet/lipgloss` — Terminal styling
- `github.com/pelletier/go-toml/v2` — Config parsing
- `gopkg.in/yaml.v3` — YAML frontmatter parsing

## Task Parsing

Tasks are `- [ ]` / `- [x]` / `- [X]` lines in any `.md` file. The parser extracts wikilinks (`[[...]]`) and tags (`#tag`) but preserves them in the description. Files with YAML frontmatter containing `type: task` + `due` + `caldav-uid` attach that metadata to the task. Each task tracks its source file path and line number for write-back.

## Obsidian Vault

Target vault: `/mnt/d/obsidian/notes`. Daily notes in `diary/` as `YYYY-MM-DD.md`. Master task list at `todo.md` in vault root. Config path is user-configurable via TOML.

## CalDAV

Uses HTTP PUT (push) and REPORT (pull) with basic auth against a CalDAV server. UID mappings stored in `~/.config/obia/sync.json`. The protocol logic mirrors the existing TypeScript plugin — same VTODO format, same status mapping (NEEDS-ACTION, IN-PROCESS, COMPLETED, CANCELLED).

## Git

- Never add a `Co-Authored-By` trailer to commits. All commits are authored solely by the user.

## Design Decisions

- Parse **all** `.md` files, not just daily notes — tasks can live anywhere in the vault
- Plain checkbox syntax only (no emoji markers, no dataview fields) — matches this vault's conventions
- Vim-style keybindings (j/k navigation) — user preference
- TOML config over YAML/JSON — human-editable with comments
- Write changes directly to `.md` files (toggle checkbox status in-place)
