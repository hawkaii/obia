# Obia — TODO

## Refactor (Phase 5)
- [ ] create ProgramContext (config, dimensions, view state, error) — pass to all components
- [ ] add state machine to root App (`modeBrowser` now, `modeChat` later)
- [ ] define Section interface + BaseModel for shared tab logic
- [ ] split app.go — extract each tab into its own section model
- [ ] centralize keybindings in `keys/keys.go` using `bubbles/key`
- [ ] replace inline logic with typed messages (`TasksLoadedMsg`, `CalDAVPushedMsg`, etc.)
- [ ] move async ops to `commands.go` — CalDAV push/pull become non-blocking
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
- [ ] add option to filter only daily note tasks
- [ ] add sorting option: most recently added (by file mtime or line position)
- [ ] first-run interactive setup (vault path, caldav creds)

## Bugs
- [ ] fix filter mode — pressing `/` appends a `/` character to the filter input

## Future (Phase 7)
- [ ] chat mode (`modeChat` in state machine, `ctrl+t` to toggle)
- [ ] voice input via Whisper API
- [ ] AI auto-linking to vault notes
- [ ] bookmark capture
- [ ] Flutter mobile app
