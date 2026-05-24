# obia

Terminal TUI for Obsidian vault task management. Go + Bubble Tea.

Scan all `.md` files in your vault, parse `- [ ]` checkboxes, manage tasks in a vim-navigable terminal UI with CalDAV sync and AI context.

## New: Hermes Integration (feat/hermes-integration)

✧ **Context tab** — session briefing on startup: last vault logs, recent GitHub commits, DSA streak, uncommitted changes  
✧ **Chat mode** — press `c` to talk to Hermes about your context  
✧ **Zero manual input** — open obia, see where you left off  

## Quick Start

```bash
go build -o obia .
./obia
```

Configure at `~/.config/obia/config.toml`.

## How It Works

```
obia TUI ──HTTP──> Hermes Context API (GCP VM)
                      │
                      ├── vault logs
                      ├── GitHub commits
                      ├── DSA progress
                      └── Hermes memory
```

## Keybindings

| Key | Action |
|-----|--------|
| `j/k` | Navigate |
| `Enter` | Toggle task |
| `a` | Add task |
| `e` | Edit task |
| `c` | **Chat mode** (Hermes AI) |
| `p` / `R` | CalDAV push / pull |
| `Tab` | Switch tabs |
| `/` | Filter |
| `q` | Quit |

## Architecture

- `internal/vault/` — file scanner, parser, writer, cache
- `internal/tui/` — Bubble Tea app, sections, forms
- `internal/caldav/` — CalDAV HTTP client (push/pull VTODOs)
- `internal/hermes/` — Hermes Context API client (new)
- `internal/config/` — TOML config
- `internal/task/` — data model

## Roadmap

| Phase | What |
|-------|------|
| 0 ✅ | Session Context Oracle + chat mode |
| 1 🔜 | Habit engine (CSES gate, streaks) |
| 2 🔜 | Two-brain bridge (personal + work vaults) |
| 3 🔜 | Anti-manual exosuit (auto git, auto push) |
| 4 🔜 | Full Hermes agent mode in chat |

## License

MIT
