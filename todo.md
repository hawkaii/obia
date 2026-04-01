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
- [ ] CalDAV: add `DESCRIPTION` (long-form body, separate from SUMMARY), `DTSTART` (start date), and time support (HH:MM on due/start fields, sends full datetime to VTODO instead of date-only) to add form and VTODO builder
- [ ] add form: upgrade Description field from single-line input to multi-line textarea
- [ ] add form: split Due into two fields — date `[YYYY-MM-DD]` + optional time `[HH:MM]`
- [ ] task file model: on push (add form or `p`), create `tasks/<uid>.md` with YAML frontmatter + title header + description; write `- [ ] [[uid|title]]` wikilink into source file (see sync.md)
- [ ] CalDAV pull (`R` key): REPORT all VTODOs, update existing task file frontmatter, create new task files + inbox entries for remote-only tasks; auto-reload after pull
- [ ] toggle linked tasks: when checkbox flipped on `[[uid|title]]` task, also update status in task file frontmatter and push to CalDAV; show error in status bar if push fails
- [ ] `p` key: open add form pre-filled with task description, on submit rewrite plain task line to `[[uid|title]]` wikilink and create task file
- [ ] task detail view: press `d` or `enter` to render `tasks/<uid>.md` content as preview overlay (see sync.md)
- [ ] open task source in Obsidian: press `o` to launch `obsidian://open?vault=...&file=...` URI — opens the note in the Obsidian app directly from Obia (works from WSL via Windows interop)

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
