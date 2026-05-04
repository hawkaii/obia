# Obia Sync Architecture

## Task File Model

When a task is added via the add form (with metadata) or pushed with `p`, Obia creates a
dedicated markdown file in the vault's task files folder (configurable, default `tasks/`),
named by the CalDAV UID for global uniqueness:

```
tasks/<caldav-uid>.md
```

### Task file format

```markdown
---
type: task
caldav-uid: 3f2a1b4c-dead-beef-1234-567890abcdef
due: 2026-04-02
status: NEEDS-ACTION
priority: 1
---

# Buy milk at 3pm

Optional longer description here (from DESCRIPTION field in VTODO)...
```

### Entry in source file (daily note / todo.md)

Uses Obsidian wikilink alias syntax so both Obia and Obsidian render the title:

```markdown
- [ ] [[3f2a1b4c-dead-beef-1234-567890abcdef|Buy milk at 3pm]]
```

The UID is the filename, the title is the display alias:
- The link is always unique (UID) and always readable (alias)
- Obsidian resolves it to the task file for backlinks, graph view, and note preview
- Obia resolves the wikilink → reads frontmatter → gets full metadata (due, status, priority)

---

## Config

```toml
[vault]
task_files_folder = "tasks"   # folder for task files, relative to vault root
inbox_file = "tasks/inbox.md" # landing zone for tasks pulled from other clients
```

---

## Sync Flow

### Push via add form (new task)

1. User fills add form (summary, description, due date, due time, target, priority, status, push toggle)
2. CalDAV UID generated
3. `tasks/` folder created if missing (if `tasks` exists as a file, use `tasks_1/` and notify)
4. `tasks/<uid>.md` created with YAML frontmatter + `# Summary` header + description body
5. `- [ ] [[<uid>|Summary]]` written into target file (daily note or todo.md)
6. VTODO pushed to CalDAV server

### Push via `p` on existing plain task

1. `p` on `- [ ] grab coffee` opens the add form pre-filled with "grab coffee" as Summary
2. User can edit Summary (becomes wikilink alias and task file title), fill other fields, submit
3. Same steps 2-6 as above
4. Source line rewritten: `- [ ] grab coffee` → `- [ ] [[<uid>|Summary]]` preserving checkbox state

### Pull (CalDAV → local) — triggered by `R`

CalDAV server is the source of truth (same model as Thunderbird and org-caldav).

1. REPORT query fetches all VTODOs from CalDAV server
2. For each VTODO:
   - If `tasks/<uid>.md` exists → update frontmatter fields (due, status, priority) with server values; append DESCRIPTION below header if present and not already there
   - If not (task created on another client) → create `tasks/<uid>.md` + append `- [ ] [[<uid>|Title]]` to inbox file
3. Checkbox state in source file stays untouched (toggle is the user's local action)
4. Task list auto-reloads after pull completes

The inbox file (configurable, default `tasks/inbox.md`) is the designated landing zone for
remote-only tasks — same pattern org-caldav uses for new org headings.

### Toggle on a linked task

When user presses space on a `[[uid|title]]` task:
1. Flips checkbox in the source file (existing behaviour)
2. Updates `status:` in `tasks/<uid>.md` frontmatter (Todo → COMPLETED, Done → NEEDS-ACTION)
3. Pushes updated status to CalDAV immediately
4. If CalDAV push fails — leaves local state as-is, shows error in status bar

### Plain tasks (no form)

`- [ ] grab coffee` stays as a plain checkbox. No task file, no CalDAV entry unless the user
explicitly presses `p` to push.

---

## sync.json

Becomes redundant once task files carry the UID in their frontmatter. Remove after migration.
Parser hydrates UID, due, status, priority by resolving wikilinks to task files.

---

## Pull Trigger

- **Now**: manual `R` key only — explicit, no surprise network activity
- **Future**: on startup (behind config flag) + background interval (needs goroutine + message passing)

---

## Start, Due, and Repeat Fields in Forms

The add/edit forms support:
- `Start date: [2026-04-01]`
- `Start time: [09:00     ]` (optional — if empty, sends date-only DTSTART)
- `Due date:   [2026-04-02]`
- `Due time:   [15:30     ]` (optional — if empty, sends date-only DUE)
- `Repeat:     [none|daily|weekly|monthly|yearly]`

Repeat values map to RRULE:
- `daily` -> `FREQ=DAILY`
- `weekly` -> `FREQ=WEEKLY`
- `monthly` -> `FREQ=MONTHLY`
- `yearly` -> `FREQ=YEARLY`

Task file frontmatter persists this as:
- `dtstart: <RFC3339>`
- `due: <RFC3339>`
- `rrule: FREQ=...`

---

## Future: Preview Panel

A split-pane or overlay that renders the content of `tasks/<uid>.md` when the cursor is on
a linked task — shows title, description body, frontmatter metadata. Activated by `d` or `enter`.
Works naturally because the task file is a real vault note the user can also open in Obsidian.
