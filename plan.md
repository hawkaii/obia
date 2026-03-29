# Obia — Implementation Plan

An interactive TUI for Obsidian vault task management, built with Go + Bubble Tea.

## v1 Scope

### Core Features
1. **Task Parsing** — Scan all `.md` files in vault for `- [ ]` / `- [x]` / `- [X]` checkbox syntax
2. **Interactive TUI** — Tabbed interface (Tasks / Today / Overdue / CalDAV) with vim-style navigation
3. **Toggle Done** — Mark tasks complete/incomplete, write changes back to source `.md` file
4. **Add Task** — Append new tasks to `todo.md` or today's daily note
5. **CalDAV Push** — Push selected tasks to CalDAV server as VTODO items
6. **CalDAV Pull** — Pull task status updates from CalDAV server

### Out of Scope (v1)
- Voice input / transcription
- AI linking / task extraction
- Bookmarks
- Flutter app
- Recurring tasks
- Tag-based filtering (stretch goal)

---

## Architecture

```
obia/
├── main.go                 # Entry point, config loading, launch TUI
├── go.mod
├── go.sum
├── config.toml.example     # Example config
│
├── internal/
│   ├── config/
│   │   └── config.go       # TOML config parsing (~/.config/obia/config.toml)
│   │
│   ├── vault/
│   │   ├── scanner.go      # Recursively find all .md files (skip .obsidian, .trash)
│   │   ├── parser.go       # Parse - [ ] tasks from markdown, track file + line offset
│   │   └── writer.go       # Write task status changes back to .md files
│   │
│   ├── task/
│   │   └── task.go         # Task data model (description, status, due, source, tags, wikilinks)
│   │
│   ├── caldav/
│   │   ├── client.go       # HTTP client (basic auth, PUT, REPORT)
│   │   ├── vtodo.go        # Build/parse VTODO iCalendar format
│   │   └── sync.go         # Push/pull logic, UID tracking
│   │
│   └── tui/
│       ├── app.go          # Root Bubble Tea model, tab switching
│       ├── tabs.go         # Tab definitions and navigation
│       ├── tasklist.go     # Task list component (filterable, scrollable)
│       ├── taskform.go     # Add task input form
│       ├── statusbar.go    # Bottom bar with keybindings help
│       └── styles.go       # Lipgloss styles and theme
│
└── README.md
```

---

## Config

`~/.config/obia/config.toml`

```toml
[vault]
path = "/mnt/d/obsidian/notes"
daily_notes_folder = "diary"
daily_notes_format = "2006-01-02"  # Go date format
default_task_file = "todo.md"

[caldav]
url = ""
username = ""
password = ""

[ui]
default_tab = "tasks"
```

---

## Task Data Model

```go
type TaskStatus int

const (
    Todo TaskStatus = iota
    Done
    Cancelled
)

type Task struct {
    Description string      // Raw text after checkbox, preserving wikilinks
    Status      TaskStatus
    Due         *time.Time  // From CalDAV frontmatter if present
    Tags        []string    // Extracted #tags
    WikiLinks   []string    // Extracted [[links]]
    Source      TaskSource
    CalDAVUID   string      // Empty if not synced
}

type TaskSource struct {
    FilePath string  // Absolute path to .md file
    Line     int     // Line number (1-indexed)
    Offset   int     // Byte offset of the checkbox in file
}
```

---

## Task Parsing Rules

1. Scan all `.md` files recursively under vault path
2. Skip directories: `.obsidian`, `.trash`, `.opencode`, `templates`, `extras`
3. Match lines: `^\s*[-*+] \[([ xX])\] (.+)$`
4. Extract status: `[ ]` = Todo, `[x]`/`[X]` = Done
5. Extract wikilinks: `\[\[([^\]]+)\]\]` — store but keep in description
6. Extract tags: `(?<!\S)#(\S+)` — store but keep in description
7. For files with YAML frontmatter containing `type: task` + `due` + `caldav-uid`, attach those to the task
8. Track file path + line number for write-back

---

## TUI Layout

```
┌─ Obia ──────────────────────────────────────────────┐
│  [Tasks]  [Today]  [Overdue]  [CalDAV]              │
├─────────────────────────────────────────────────────┤
│                                                     │
│  Task list (scrollable, filterable)                 │
│  Each row: [status] description          source.md  │
│                                                     │
├─────────────────────────────────────────────────────┤
│  ↑/k ↓/j navigate  enter: toggle  p: push caldav   │
│  /: filter  a: add task  tab: switch tab  q: quit   │
└─────────────────────────────────────────────────────┘
```

### Tab Views
- **Tasks** — All open tasks across vault, grouped by source file
- **Today** — Tasks from `diary/<today>.md` + tasks with `due` = today
- **Overdue** — Tasks with `due` date before today
- **CalDAV** — Tasks that have a `caldav-uid` (synced tasks), show remote status

### Key Bindings
| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `tab` | Next tab |
| `shift+tab` | Previous tab |
| `enter` | Toggle task done/undone |
| `p` | Push selected task to CalDAV |
| `P` | Pull all from CalDAV |
| `a` | Add new task |
| `/` | Search/filter tasks |
| `esc` | Clear filter / cancel |
| `q` | Quit |

---

## Implementation Phases

### Phase 1 — Scaffold & Config
- [ ] `go mod init github.com/hawkaii/obia`
- [ ] TOML config loading with defaults
- [ ] First-run interactive setup (vault path, caldav creds)
- [ ] Skeleton Bubble Tea app that launches and quits

### Phase 2 — Vault Scanning & Task Parsing
- [ ] Recursive `.md` file scanner with directory exclusions
- [ ] Markdown task parser (checkbox regex, wikilink/tag extraction)
- [ ] YAML frontmatter parser for CalDAV task notes
- [ ] Task data model with source tracking

### Phase 3 — TUI Task List
- [ ] Task list component with scrolling
- [ ] Render tasks with status icon, description, source file
- [ ] Vim-style navigation (j/k)
- [ ] Tab bar (Tasks tab only initially)
- [ ] Status bar with keybinding hints

### Phase 4 — Task Actions
- [ ] Toggle done: write `[x]` / `[ ]` back to source file
- [ ] Add task: input form, append to `todo.md` or today's daily note
- [ ] Search/filter tasks by text

### Phase 5 — CalDAV Integration
- [ ] Port CalDAV client from TypeScript plugin (HTTP PUT/REPORT with basic auth)
- [ ] Build VTODO from task data
- [ ] Push: generate UID, send to server, store UID mapping locally
- [ ] Pull: fetch remote statuses, update task display
- [ ] UID mapping stored in `~/.config/obia/sync.json`

### Phase 6 — Tabs & Views
- [ ] Today tab (filter by daily note date + due date)
- [ ] Overdue tab (filter by due < today)
- [ ] CalDAV tab (filter by has caldav-uid)
- [ ] Tab switching with tab/shift+tab

---

## Dependencies

```
github.com/charmbracelet/bubbletea     # TUI framework
github.com/charmbracelet/bubbles       # TUI components (list, textinput, viewport)
github.com/charmbracelet/lipgloss      # Styling
github.com/pelletier/go-toml/v2        # TOML config parsing
gopkg.in/yaml.v3                       # YAML frontmatter parsing
```

---

## CalDAV Port Notes

Source: `~/code/projects/caldev-sync-plugin/caldav-sync/src/caldav.ts`

What to port:
- `buildVTodo(uid, summary, due)` → `vtodo.go`
- `pushTodo(settings, uid, icsData)` → `client.go` (swap `requestUrl` for `net/http`)
- `pullTodos(settings)` → `client.go` (REPORT method + XML parsing)
- `generateUID()` → use Go's `github.com/google/uuid`
- Basic auth header generation
- iCal date formatting

The TypeScript CalDAV code is ~142 lines. The Go port should be roughly the same size.

---

## Future (post-v1)

- Voice input via Whisper API → transcribe → extract tasks
- AI auto-linking (match wikilinks to existing vault notes)
- Bookmark capture (URL + title → save to vault)
- Flutter mobile app sharing the same vault
- Background file watcher (fsnotify) for live task updates
- Recurring task support (RRULE parsing)
- Priority support (emoji markers or custom syntax)
