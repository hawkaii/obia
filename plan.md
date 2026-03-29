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
├── main.go                     # Entry point, config loading, launch TUI
├── go.mod
├── go.sum
├── config.toml.example
│
├── internal/
│   ├── config/
│   │   └── config.go           # TOML config parsing (~/.config/obia/config.toml)
│   │
│   ├── vault/
│   │   ├── scanner.go          # Recursively find all .md files (skip .obsidian, .trash)
│   │   ├── parser.go           # Parse - [ ] tasks from markdown, track file + line offset
│   │   └── writer.go           # Write task status changes back to .md files
│   │
│   ├── task/
│   │   └── task.go             # Task data model (status, description, due, source, tags, wikilinks)
│   │
│   ├── caldav/
│   │   ├── client.go           # HTTP client (basic auth, PUT, REPORT)
│   │   ├── vtodo.go            # Build/parse VTODO iCalendar format
│   │   └── sync.go             # Push/pull logic, UID tracking
│   │
│   └── tui/
│       ├── app.go              # Root model: message dispatch, tab routing, view composition
│       ├── messages.go         # All typed messages (TasksLoadedMsg, CalDAVPushedMsg, etc.)
│       ├── commands.go         # tea.Cmd factories for async ops (load tasks, push caldav)
│       ├── styles.go           # Lipgloss styles and theme
│       │
│       ├── context/
│       │   └── context.go      # ProgramContext: config, dimensions, view state, error
│       │
│       ├── keys/
│       │   └── keys.go         # KeyMap definitions using bubbles/key, view-aware help
│       │
│       └── components/
│           ├── section/
│           │   └── section.go  # Section interface + BaseModel (shared cursor, search, loading)
│           ├── tasksection/
│           │   └── model.go    # Task list section implementing Section interface
│           ├── tabs/
│           │   └── tabs.go     # Tab bar rendering and navigation
│           └── statusbar/
│               └── statusbar.go # Bottom bar with context-aware keybinding hints
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

### Phase 1 — Scaffold & Config [DONE]
- [x] `go mod init github.com/hawkaii/obia`
- [x] TOML config loading with defaults
- [x] Skeleton Bubble Tea app that launches and quits

### Phase 2 — Vault Scanning & Task Parsing [DONE]
- [x] Recursive `.md` file scanner with directory exclusions
- [x] Markdown task parser (checkbox regex, wikilink/tag extraction)
- [x] YAML frontmatter parser for CalDAV task notes
- [x] Task data model with source tracking
- [x] Smoke tested against real vault: 226 tasks (194 open, 32 done)

### Phase 3 — TUI Task List [DONE]
- [x] Task list component with scrolling
- [x] Render tasks with status icon, description, source file
- [x] Vim-style navigation (j/k)
- [x] Tab bar with all 4 tabs (Tasks/Today/Overdue/CalDAV)
- [x] Status bar with keybinding hints
- [x] Toggle done: write `[x]` / `[ ]` back to source file
- [x] Add task: input form, append to `todo.md`
- [x] Search/filter tasks by text

### Phase 4 — CalDAV Integration [DONE]
- [x] Port CalDAV client from TypeScript plugin (HTTP PUT/REPORT with basic auth)
- [x] Build VTODO from task data
- [x] Push: generate UID, send to server, store UID mapping locally
- [x] Pull: fetch remote statuses, update task display
- [x] UID mapping stored in `~/.config/obia/sync.json`
- [x] Tests for VTODO builder, parser, and writer

### Phase 5 — Refactor: Align with gh-dash Architecture

Studied gh-dash codebase. Key patterns to adopt:

#### 5a — ProgramContext
- [ ] Create `internal/tui/context/context.go`
- [ ] Holds: config, screen dimensions, current view, error state, styles
- [ ] All components receive `*context.ProgramContext` instead of raw config
- [ ] Central update via `UpdateProgramContext()` on resize/view change

#### 5b — Section Interface & Split app.go
- [ ] Define `Section` interface: `Update()`, `View()`, `GetId()`, `GetType()`, `NumRows()`, `CurrRow()`
- [ ] Create `internal/tui/components/section/` with base model (shared fields: ctx, cursor, loading, search)
- [ ] Extract task list into `internal/tui/components/tasksection/` implementing Section
- [ ] Each tab becomes a section instance with its own filter logic
- [ ] Root `app.go` only dispatches messages to active section

#### 5c — Keybinding System
- [ ] Create `internal/tui/keys/keys.go` with `KeyMap` struct using `charmbracelet/bubbles/key`
- [ ] Define bindings at package level, not inline strings
- [ ] View-aware help text (different bindings shown per tab)
- [ ] Support future custom keybindings from config

#### 5d — Typed Messages
- [ ] Replace inline logic with typed messages: `TasksLoadedMsg`, `TaskToggledMsg`, `TaskAddedMsg`, `CalDAVPushedMsg`, `CalDAVPulledMsg`, `ErrorMsg`
- [ ] Async operations return `tea.Cmd` that emit messages
- [ ] CalDAV push/pull become non-blocking commands

#### 5e — Data Layer Separation
- [ ] Vault parsing stays in `internal/vault/` (already clean)
- [ ] TUI never calls vault functions directly — goes through commands that return messages
- [ ] Add task caching: only re-parse changed files (track mtime like VaultMate does)

### Phase 6 — Features & Bug Fixes
- [ ] Fix filter bug: `/` keypress appends `/` character to filter input
- [ ] Add option to filter only daily note tasks
- [ ] Add sorting option: most recently added (by file mtime or line position)
- [ ] First-run interactive setup (vault path, caldav creds)

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
