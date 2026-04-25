# Obia — TODO

## Refactor (Phase 5)

- [x] create ProgramContext (config, dimensions, view state, error) — pass to all components
- [x] add state machine to root App (`modeBrowser` now, `modeChat` later)
- [x] define Section interface + BaseModel for shared tab logic
- [x] split app.go — extract each tab into its own section model
- [x] centralize keybindings in `keys/keys.go` using `bubbles/key`
- [x] replace inline logic with typed messages (`TasksLoadedMsg`, `CalDAVPushedMsg`, etc.)
- [x] move async ops to `commands.go` — CalDAV push/pull become non-blocking
- [ ] add mtime-based task caching (only re-parse changed files)

## CLI Flags (Phase 5.5)

Add Cobra for CLI framework, then implement:

### Must-have

- [ ] `--help` / `-h` — usage info and examples
- [ ] `--version` / `-v` — version + commit + build date + OS/arch
- [ ] `--config` / `-c` — override config file path (useful for testing or multiple vaults)
- [ ] `--vault` / `-V` — override vault path directly without editing config
- [ ] `--debug` — log to `debug.log` (can't print to stdout in a TUI)

### Nice-to-have

- [ ] `--tab` / `-t` — launch into a specific tab (`obia --tab today`)
- [ ] `--no-tui` — print tasks to stdout as plain text (for scripting, piping, cron jobs, status bars)

## Features (Phase 6)

- [ ] add "Daily" tab — shows all tasks from `diary/*.md` files across all dates (not just today)
- [x] fuzzy search with `sahilm/fuzzy` — replace exact substring matching in `/` filter
- [x] grouped view: tasks grouped by file with header separators, toggle with `v` between flat/grouped
- [x] add sorting option: most recently added (by file mtime or line position)
- [ ] first-run interactive setup (vault path, caldav creds)
- [x] task add form: summary, target file (fuzzy picker with ctrl+x chords), due date + time, description, priority, status, push toggle
- [x] task edit form (`e` key): edit any task in-place; plain tasks auto-upgrade to linked tasks when metadata is set
- [x] CalDAV push form: set due date, priority, and status before pushing (interactive overlay)
- [x] smart task add: route new tasks to today's daily note or default file (`add_task_target` config)
- [x] CalDAV auto-push: automatically push new tasks on add (`auto_push` config flag)
- [x] CalDAV DESCRIPTION field: stored in task file body, sent in VTODO, pulled from remote
- [x] due date + time fields: date `[YYYY-MM-DD]` + optional time `[HH:MM]` in add/edit forms
- [x] task file model: `tasks/<uid>.md` with YAML frontmatter + title header + description body; source line rewritten to `[[uid|title]]` wikilink
- [x] CalDAV pull (`R` key): REPORT all VTODOs, update existing task file frontmatter, create new task files + inbox entries for remote-only tasks; auto-reload after pull
- [x] toggle linked tasks: checkbox flip updates task file status + pushes to CalDAV; error shown in status bar if push fails
- [x] `p` key: open add form pre-filled with task description, on submit rewrite plain task line to `[[uid|title]]` wikilink and create task file
- [ ] add right-side stats overview panel (OpenCode-style) in browser view
  - implement stats aggregator: total/open/done/done%, overdue/today/next 7d/no due, CalDAV linked/unlinked, top tags/files
  - responsive layout: auto-hide stats panel on narrow terminals (< 110 cols)
  - add toggle key for stats panel visibility (e.g., `s`)
  - stats computed from all vault tasks (global) + current tab counts
- [ ] task detail view: press `d` to render `tasks/<uid>.md` content as preview overlay (see sync.md)
- [ ] open task source in Obsidian: press `o` to launch `obsidian://open?vault=...&file=...` URI — opens the note in the Obsidian app directly from Obia (works from WSL via Windows interop)
- [ ] add form: upgrade Description field from single-line input to multi-line textarea
- [x] CalDAV: add `DTSTART` (start date) support
- [x] add repeat option in add/edit forms (`none`, `daily`, `weekly`, `monthly`, `yearly`) with RRULE sync

## Bugs

- [x] fix filter mode — pressing `/` appends a `/` character to the filter input

## Google Integration (Phase 6.5) — not final, exploring options

- [ ] Google Tasks/Calendar integration via `gws` CLI

## Future (Phase 7)

- [ ] chat mode (`modeChat` in state machine, `ctrl+t` to toggle)
- [ ] voice input
- [ ] AI auto-linking to vault notes
- [ ] bookmark capture
- [ ] Flutter mobile app
