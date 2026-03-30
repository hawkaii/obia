# Obia

<p>
    <img src="https://img.shields.io/badge/go-%3E%3D1.22-blue?logo=go" alt="Go Version">
    <a href="https://github.com/hawkaii/obia/actions"><img src="https://img.shields.io/badge/build-passing-brightgreen" alt="Build Status"></a>
    <a href="https://github.com/hawkaii/obia/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue" alt="License"></a>
</p>

Your Obsidian vault, in the terminal. A fast, interactive TUI for managing tasks across your entire vault — no Obsidian needed.

<!--
<p>
    <img src="./assets/demo.gif" width="100%" alt="Obia demo">
</p>
-->

Obia scans every `.md` file in your [Obsidian](https://obsidian.md) vault, pulls out `- [ ]` tasks, and lets you browse, filter, toggle, and sync them from your terminal. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

---

## Features

- **All your tasks, one place** — Parses `- [ ]` / `- [x]` from every markdown file in your vault
- **Tabbed views** — Tasks, Today, Overdue, CalDAV — switch with <kbd>Tab</kbd>
- **Vim-style navigation** — <kbd>j</kbd>/<kbd>k</kbd> to move, <kbd>g</kbd>/<kbd>G</kbd> for top/bottom
- **Toggle done** — Hit <kbd>Enter</kbd> to check/uncheck, writes back to the `.md` file instantly
- **Smart task add** — Press <kbd>a</kbd> to add a task; routes to today's daily note or default file based on config
- **Fuzzy search** — <kbd>/</kbd> to filter with fuzzy matching across all task descriptions
- **Grouped view** — Press <kbd>:</kbd> to toggle between flat list and tasks grouped by source file
- **CalDAV sync** — Push tasks to any CalDAV server with <kbd>p</kbd> (Radicale, Nextcloud, iCloud, etc.)
- **CalDAV push form** — Set due date, priority, and status in an interactive overlay before pushing
- **CalDAV auto-push** — Optionally push new tasks to CalDAV automatically on add
- **Wikilinks & tags** — Extracts `[[links]]` and `#tags` from task descriptions
- **Fast** — Scans 800+ files in under a second

---

## Install

```bash
go install github.com/hawkaii/obia@latest
```

Or build from source:

```bash
git clone https://github.com/hawkaii/obia.git
cd obia
go build -o obia .
```

---

## Getting Started

### 1. Create a config

Obia looks for its config at `~/.config/obia/config.toml`. Create one:

```bash
mkdir -p ~/.config/obia
```

```toml
[vault]
path = "/path/to/your/obsidian/vault"
daily_notes_folder = "diary"
daily_notes_format = "2006-01-02"
default_task_file = "todo.md"
add_task_target = "daily"   # "daily" → today's note, "default" → default_task_file

[caldav]
url = ""
username = ""
password = ""
auto_push = false           # push new tasks to CalDAV automatically on add

[ui]
default_tab = "tasks"
```

### 2. Run it

```bash
obia
```

That's it. You'll see all your open tasks across the vault.

---

## Keybindings

| Key | Action |
|-----|--------|
| <kbd>j</kbd> / <kbd>↓</kbd> | Move down |
| <kbd>k</kbd> / <kbd>↑</kbd> | Move up |
| <kbd>g</kbd> / <kbd>G</kbd> | Jump to top / bottom |
| <kbd>Tab</kbd> / <kbd>Shift+Tab</kbd> | Switch tabs |
| <kbd>Enter</kbd> | Toggle task done/undone |
| <kbd>a</kbd> | Add new task (routes to daily note or default file) |
| <kbd>/</kbd> | Fuzzy search / filter |
| <kbd>:</kbd> | Toggle flat / grouped-by-file view |
| <kbd>p</kbd> | Push task to CalDAV (opens form for due/priority/status) |
| <kbd>r</kbd> | Reload vault |
| <kbd>Esc</kbd> | Clear filter / cancel |
| <kbd>q</kbd> | Quit |

---

## Tabs

- **Tasks** — All open tasks across your vault
- **Today** — Tasks from today's daily note + tasks due today
- **Overdue** — Tasks past their due date
- **CalDAV** — Tasks synced with your CalDAV server

---

## CalDAV Sync

Obia can push tasks to any CalDAV-compatible server. Fill in the `[caldav]` section of your config:

```toml
[caldav]
url = "https://your-server.com/radicale/user/calendar/"
username = "you"
password = "secret"
```

Then press <kbd>p</kbd> on any task to open the push form. You can edit the summary, set a due date, choose priority (none / 1 / 5 / 9), and set CalDAV status (NEEDS-ACTION, IN-PROCESS, COMPLETED, CANCELLED) before pushing.

UID mappings are stored in `~/.config/obia/sync.json` and hydrated back into tasks on every load — your vault markdown stays clean.

To push automatically whenever you add a task, set `auto_push = true` in `[caldav]`.

---

## Task Format

Obia looks for standard Obsidian checkbox syntax in any `.md` file:

```markdown
- [ ] open task
- [x] completed task
- [ ] task with [[wikilink]] and #tag
```

Files with YAML frontmatter are also supported for CalDAV metadata:

```yaml
---
type: task
title: "Deploy the fix"
due: 2026-04-01
caldav-uid: abc-123-def
---
```

---

## How It Works

```
~/.config/obia/config.toml
         │
         ▼
   ┌──────────┐     ┌─────────────┐
   │  Config   │────▶│ Vault       │
   │  Loader   │     │ Scanner     │──▶ finds all .md files
   └──────────┘     └─────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │ Task Parser │──▶ extracts - [ ] lines
                    └─────────────┘
                           │
                           ▼
                    ┌─────────────┐     ┌──────────┐
                    │ Bubble Tea  │◀───▶│ CalDAV   │
                    │ TUI         │     │ Client   │
                    └─────────────┘     └──────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │ Writer      │──▶ toggles checkboxes in .md files
                    └─────────────┘
```

---

## Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) — Text input components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Terminal styling
- [go-toml](https://github.com/pelletier/go-toml) — Config parsing
- [sahilm/fuzzy](https://github.com/sahilm/fuzzy) — Fuzzy search

---

## Roadmap

- [ ] CLI flags (Cobra) — `--vault`, `--config`, `--debug`, `--no-tui`
- [ ] Daily tab — all tasks from `diary/*.md` across all dates
- [ ] Task detail view — `d` to see full metadata (due, tags, source, CalDAV UID)
- [ ] First-run setup wizard
- [ ] mtime-based task caching (skip unchanged files)
- [ ] Chat mode (`ctrl+t` to toggle, `modeChat` in state machine)
- [ ] Voice input via Whisper API
- [ ] AI auto-linking to vault notes
- [ ] Flutter mobile app

---

## License

[MIT](LICENSE)

---

<p>
    <strong>Obia</strong> — your vault, your terminal, your tasks.<br>
    <sub>Named after obsidian. Built for people who live in the terminal.</sub>
</p>
