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
- [x] grouped view: tasks grouped by file with header separators, toggle with `:` between flat/grouped
- [x] add sorting option: most recently added (by file mtime or line position)
- [ ] first-run interactive setup (vault path, caldav creds)
- [ ] task add form: support setting due date/time, tags, and target file (not just todo.md)
- [x] CalDAV push form: set due date, priority, and status before pushing (interactive overlay)
- [x] smart task add: route new tasks to today's daily note or default file (`add_task_target` config)
- [x] CalDAV auto-push: automatically push new tasks on add (`auto_push` config flag)
- [ ] task detail view: press `d` or `enter` to see full task metadata (due, tags, source, CalDAV UID)

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
