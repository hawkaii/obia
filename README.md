# Obia

<p>
    <img src="https://img.shields.io/badge/go-%3E%3D1.24-blue?logo=go" alt="Go Version">
    <a href="https://github.com/hawkaii/obia/actions"><img src="https://img.shields.io/badge/build-passing-brightgreen" alt="Build Status"></a>
    <a href="https://github.com/hawkaii/obia/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue" alt="License"></a>
</p>

Your Obsidian vault, in the terminal. A fast, interactive TUI for managing tasks across your entire vault — no Obsidian needed.

<!--
<p>
    <img src="./assets/demo.gif" width="100%" alt="Obia demo">
</p>
-->

Obia scans every `.md` file in your [Obsidian](https://obsidian.md) vault, pulls out `- [ ]` tasks, and lets you browse, filter, toggle, add, edit, and sync them from your terminal. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

---

## Features

- **All your tasks, one place** — Parses `- [ ]` / `- [x]` from every markdown file in your vault
- **Tabbed views** — Tasks, Today, Overdue, CalDAV — switch with <kbd>Tab</kbd>
- **Vim-style navigation** — <kbd>j</kbd>/<kbd>k</kbd> to move, <kbd>g</kbd>/<kbd>G</kbd> for top/bottom
- **Toggle done** — Hit <kbd>Enter</kbd> to check/uncheck, writes back to the `.md` file instantly
- **Rich task add** — Press <kbd>a</kbd> to open a form: summary, target file (fuzzy picker), due date, time, description, priority, status, optional CalDAV push
- **Edit tasks** — Press <kbd>e</kbd> to edit any task; plain tasks auto-upgrade to linked tasks when metadata is set
- **Fuzzy search** — <kbd>/</kbd> to filter with fuzzy matching across all task descriptions
- **Grouped view** — Press <kbd>v</kbd> to toggle between flat list and tasks grouped by source file
- **Task files** — Linked tasks stored as `tasks/<uid>.md` with YAML frontmatter; wikilinks keep your vault clean
- **CalDAV sync** — Push tasks to any CalDAV server with <kbd>p</kbd> (Radicale, Nextcloud, iCloud, Tasks.org, etc.)
- **CalDAV pull** — Press <kbd>R</kbd> to pull all VTODOs from your CalDAV server; new remote tasks land in your inbox
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
add_task_target = "daily"       # default target when pressing `a`: "daily" or "default"
task_files_folder = "tasks"     # folder where task files (tasks/<uid>.md) are stored
inbox_file = "tasks/inbox.md"   # file where pulled CalDAV tasks are appended
extra_targets = [               # additional target files shown in the add/edit form picker
    "projects/work.md",
    "projects/personal.md",
]

[caldav]
url = ""
username = ""
password = ""
auto_push = false               # push new tasks to CalDAV automatically on add
```

### 2. Run it

```bash
obia
```

That's it. You'll see all your open tasks across the vault.

---

## iOS Sync via iCloud (Reminders)

If you want tasks pushed from Obia to appear in the **Reminders** app on your iPhone, point Obia at iCloud's CalDAV server — no third-party server needed.

### 1. Generate an app-specific password

1. Go to [appleid.apple.com](https://appleid.apple.com) and sign in
2. Under **Sign-In and Security**, choose **App-Specific Passwords → Generate**
3. Name it something like `obia` and copy the generated password (`xxxx-xxxx-xxxx-xxxx`)

### 2. Add CalDAV config

```toml
[caldav]
url      = "https://caldav.icloud.com"
username = "your@icloud.com"
password = "xxxx-xxxx-xxxx-xxxx"   # app-specific password, NOT your Apple ID password
```

### 3. Sync

| Action | How |
|--------|-----|
| Push a task to Reminders | Press <kbd>p</kbd> on any task, fill the form, submit |
| Push on add | Set `auto_push = true` in config, or toggle **Push?** in the add form |
| Pull tasks from Reminders | Press <kbd>R</kbd> — new remote tasks land in your inbox file |
| Toggle done syncs too | Checking/unchecking a linked task updates Reminders automatically |

Tasks appear in the default **Reminders** list on iOS. They will **not** appear in the Calendar app — Reminders is Apple's CalDAV task store.

---

## Keybindings

| Key | Action |
|-----|--------|
| <kbd>j</kbd> / <kbd>↓</kbd> | Move down |
| <kbd>k</kbd> / <kbd>↑</kbd> | Move up |
| <kbd>g</kbd> / <kbd>G</kbd> | Jump to top / bottom |
| <kbd>Tab</kbd> / <kbd>Shift+Tab</kbd> | Switch tabs |
| <kbd>Enter</kbd> | Toggle task done/undone |
| <kbd>a</kbd> | Add new task (opens rich form) |
| <kbd>e</kbd> | Edit task (summary, due date/time, description, priority, status) |
| <kbd>p</kbd> | Upgrade plain task to linked task (opens form pre-filled) |
| <kbd>R</kbd> | Pull all tasks from CalDAV server |
| <kbd>/</kbd> | Fuzzy search / filter |
| <kbd>v</kbd> | Toggle flat / grouped-by-file view |
| <kbd>r</kbd> | Reload vault |
| <kbd>Esc</kbd> | Clear filter / cancel form |
| <kbd>q</kbd> | Quit |

---

## Tabs

- **Tasks** — All open tasks across your vault
- **Today** — Tasks due today
- **Overdue** — Tasks past their due date
- **CalDAV** — Tasks synced with your CalDAV server

---

## Task Model

Obia has two kinds of tasks:

### Plain tasks
Standard Obsidian checkbox syntax in any `.md` file:
```markdown
- [ ] buy groceries
- [x] finished task
- [ ] task with [[wikilink]] and #tag
```

### Linked tasks
When you add metadata (due date, priority, etc.) or push to CalDAV, Obia creates a task file and rewrites the checkbox as a wikilink alias:

```markdown
- [ ] [[3f8a1b2c-...|buy groceries]]
```

The task file at `tasks/<uid>.md` stores all the metadata:

```markdown
---
type: task
caldav-uid: 3f8a1b2c-...
due: 2026-04-02T09:00:00Z
priority: 5
status: NEEDS-ACTION
---

# buy groceries

Optional longer description here.
```

This keeps your vault markdown clean while preserving full CalDAV metadata.

---

## CalDAV Sync

Obia syncs with any CalDAV-compatible server (Radicale, Nextcloud, iCloud, Tasks.org, etc.).

Fill in the `[caldav]` section of your config:

```toml
[caldav]
url = "https://your-server.com/radicale/user/calendar/"
username = "you"
password = "secret"
```

### Push
Press <kbd>p</kbd> on a plain task to open the form — set due date, time, description, priority, and status, then push. The task line is rewritten to `[[uid|title]]` and a task file is created.

When adding a new task with <kbd>a</kbd>, toggle the **Push?** field in the form to push immediately.

### Pull
Press <kbd>R</kbd> to fetch all VTODOs from the server:
- Existing task files are updated (due date, status, priority)
- New remote tasks get a task file created and a `- [ ] [[uid|title]]` line appended to your inbox file (`tasks/inbox.md` by default)

### Toggle sync
Toggling a linked task (checking/unchecking) automatically updates the task file status and pushes the change to CalDAV.

### Auto-push
Set `auto_push = true` in `[caldav]` to push every new task automatically on add.

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
                    │             │──▶ resolves [[uid|alias]] wikilinks
                    │             │──▶ hydrates due/status/priority from task files
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
                    │             │──▶ creates/updates tasks/<uid>.md
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

## Works With

Obia works with **any folder of markdown files** — not just Obsidian. It only needs standard checkbox syntax and `.md` files.

| Tool | Compatibility |
|------|--------------|
| [Obsidian](https://obsidian.md) | Full |
| [Logseq](https://logseq.com) | Full — same `- [ ]` syntax, wikilinks, YAML frontmatter |
| [Foam](https://foambubble.github.io/foam/) | Full — VS Code extension, same conventions |
| [Zettlr](https://zettlr.com) | Full — standard markdown + frontmatter |
| [Dendron](https://www.dendron.so) | Full — VS Code extension, hierarchical `.md` files |
| Any folder of `.md` files | Full — VS Code, Neovim, iA Writer, etc. |
| [Joplin](https://joplinapp.org) | Partial — only with filesystem sync enabled |
| Org-mode / Roam / Notion | ✗ — different formats |

Point `vault.path` in your config at any markdown directory and Obia will scan it.

---

## Roadmap

- [ ] Task detail view — press `d` to render `tasks/<uid>.md` as a preview overlay
- [ ] Open in Obsidian — press `o` to launch `obsidian://open?vault=...&file=...`
- [ ] CLI flags (Cobra) — `--vault`, `--config`, `--debug`, `--no-tui`
- [ ] Daily tab — all tasks from `diary/*.md` across all dates
- [ ] First-run setup wizard
- [ ] mtime-based task caching (skip unchanged files)
- [ ] Multi-line description textarea

---

## Sponsoring

If you find Obia useful, consider supporting the project:

[![Sponsor](https://img.shields.io/badge/sponsor-%E2%9D%A4-ea4aaa?logo=github)](https://github.com/sponsors/hawkaii)

## Discord

Have questions or want to chat? Join the community:

[![Discord](https://img.shields.io/badge/discord-join-5865F2?logo=discord&logoColor=white)](https://discord.gg/5YfG9nS74h)

## License

[MIT](LICENSE)

---

<p>
    <strong>Obia</strong> — your vault, your terminal, your tasks.<br>
    <sub>Named after obsidian. Built for people who live in the terminal.</sub>
</p>
