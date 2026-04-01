package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/hawkaii/obia/internal/caldav"
	"github.com/hawkaii/obia/internal/config"
	"github.com/hawkaii/obia/internal/task"
	"github.com/hawkaii/obia/internal/vault"
)

// LoadTasksCmd scans the vault and returns all parsed tasks.
func LoadTasksCmd(vaultPath, taskFilesFolder string) tea.Cmd {
	return func() tea.Msg {
		tasks, err := vault.ParseAllTasks(vaultPath, taskFilesFolder)
		return TasksLoadedMsg{Tasks: tasks, Err: err}
	}
}

// ToggleTaskCmd writes the toggled checkbox back to the source file.
// For linked tasks, also updates status in the task file and pushes to CalDAV.
func ToggleTaskCmd(t *task.Task, caldavCfg config.CalDAV) tea.Cmd {
	return func() tea.Msg {
		if err := vault.ToggleTask(t); err != nil {
			return TaskToggledMsg{Task: t, Err: err}
		}

		if t.LinkedTaskFile != "" {
			newStatus := "NEEDS-ACTION"
			if t.Status == task.Done {
				newStatus = "COMPLETED"
			}
			_ = vault.UpdateTaskFileStatus(t.LinkedTaskFile, newStatus)

			if caldavCfg.URL != "" {
				_, pushErr := caldav.PushTask(caldavCfg, t, t.Due, 0, newStatus, "")
				if pushErr != nil {
					return TaskToggledMsg{Task: t, CalDAVErr: pushErr}
				}
			}
		}

		return TaskToggledMsg{Task: t}
	}
}

// AddTaskCmd appends a plain task with no task file.
func AddTaskCmd(filePath, description string) tea.Cmd {
	return func() tea.Msg {
		err := vault.AppendTask(filePath, description)
		return TaskAddedMsg{Description: description, Err: err}
	}
}

// AddTaskWithMetaCmd creates a task file, writes [[uid|title]] into the target file,
// and optionally pushes to CalDAV.
func AddTaskWithMetaCmd(
	filePath, summary, description string,
	due *time.Time,
	priority int,
	status string,
	push bool,
	cfg config.Config,
) tea.Cmd {
	return func() tea.Msg {
		taskFilesDir, _ := vault.EnsureTaskFolder(cfg.Vault.Path, cfg.Vault.TaskFilesFolder)

		uid := uuid.New().String()
		var pushErr error

		if push && cfg.CalDAV.URL != "" {
			tmpTask := &task.Task{Description: summary, CalDAVUID: uid}
			uid2, err := caldav.PushTask(cfg.CalDAV, tmpTask, due, priority, status, description)
			if err != nil {
				pushErr = err
			} else {
				uid = uid2
			}
		}

		if err := vault.CreateTaskFile(taskFilesDir, uid, summary, description, due, priority, status); err != nil {
			return TaskAddedMsg{Description: summary, Err: err}
		}

		_, err := vault.AppendTaskAt(filePath, "[["+uid+"|"+summary+"]]")
		if err != nil {
			return TaskAddedMsg{Description: summary, Err: err}
		}

		if pushErr != nil {
			return TaskAddedMsg{Description: summary, AutoPushErr: pushErr}
		}
		if push {
			return TaskAddedMsg{Description: summary, AutoPushUID: uid}
		}
		return TaskAddedMsg{Description: summary}
	}
}

// LinkExistingTaskCmd converts a plain task into a linked task:
// creates a task file, rewrites the source line to [[uid|title]], optionally pushes.
func LinkExistingTaskCmd(
	t *task.Task,
	summary, description string,
	due *time.Time,
	priority int,
	status string,
	push bool,
	cfg config.Config,
) tea.Cmd {
	return func() tea.Msg {
		taskFilesDir, _ := vault.EnsureTaskFolder(cfg.Vault.Path, cfg.Vault.TaskFilesFolder)

		uid := uuid.New().String()
		var pushErr error

		if push && cfg.CalDAV.URL != "" {
			pushTask := &task.Task{Description: summary, CalDAVUID: uid}
			uid2, err := caldav.PushTask(cfg.CalDAV, pushTask, due, priority, status, description)
			if err != nil {
				pushErr = err
			} else {
				uid = uid2
			}
		}

		if err := vault.CreateTaskFile(taskFilesDir, uid, summary, description, due, priority, status); err != nil {
			return TaskAddedMsg{Description: summary, Err: err}
		}

		if err := vault.RewriteTaskLine(t.Source.FilePath, t.Source.Line, uid, summary); err != nil {
			return TaskAddedMsg{Description: summary, Err: err}
		}

		if pushErr != nil {
			return TaskAddedMsg{Description: summary, AutoPushErr: pushErr}
		}
		if push {
			return TaskAddedMsg{Description: summary, AutoPushUID: uid}
		}
		return TaskAddedMsg{Description: summary}
	}
}

// PullCalDAVCmd fetches all VTODOs from the server and syncs local task files.
func PullCalDAVCmd(cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		if cfg.CalDAV.URL == "" {
			return PullCalDAVMsg{}
		}

		todos, err := caldav.PullTodos(cfg.CalDAV)
		if err != nil {
			return PullCalDAVMsg{Err: err}
		}

		taskFilesDir, notify := vault.EnsureTaskFolder(cfg.Vault.Path, cfg.Vault.TaskFilesFolder)
		inboxPath := filepath.Join(cfg.Vault.Path, cfg.Vault.InboxFile)

		updated, created, fileErrs := 0, 0, 0

		for _, todo := range todos {
			if todo.UID == "" {
				continue
			}
			taskFile := filepath.Join(taskFilesDir, todo.UID+".md")

			if _, statErr := os.Stat(taskFile); statErr == nil {
				if err := vault.UpdateTaskFileFrontmatter(taskFile, todo.Due, todo.Status, todo.Priority); err != nil {
					fileErrs++
				} else {
					updated++
				}
			} else {
				if err := vault.CreateTaskFile(taskFilesDir, todo.UID, todo.Summary, todo.Description, todo.Due, todo.Priority, todo.Status); err != nil {
					fileErrs++
					continue
				}
				if err := vault.AppendToFile(inboxPath, "- [ ] [["+todo.UID+"|"+todo.Summary+"]]"); err != nil {
					fileErrs++
				} else {
					created++
				}
			}
		}

		msg := PullCalDAVMsg{Updated: updated, Created: created, Notify: notify}
		if fileErrs > 0 {
			msg.Notify = fmt.Sprintf("%d file write error(s)", fileErrs)
			if notify != "" {
				msg.Notify = notify + " · " + msg.Notify
			}
		}
		return msg
	}
}

// PushCalDAVCmd pushes an existing task to CalDAV via the push form.
func PushCalDAVCmd(cfg config.CalDAV, t *task.Task) tea.Cmd {
	return func() tea.Msg {
		uid, err := caldav.PushTask(cfg, t, nil, 0, "", "")
		return CalDAVPushedMsg{Task: t, UID: uid, Err: err}
	}
}

// EditTaskCmd saves edits to a task.
//
// Plain task, no metadata → rewrite description in-place.
// Plain task, metadata set → upgrade to linked task (creates task file + wikilink).
// Linked task → update task file frontmatter, title, body; push to CalDAV if UID set.
func EditTaskCmd(
	t *task.Task,
	newSummary, body string,
	due *time.Time,
	priority int,
	status string,
	push bool,
	cfg config.Config,
) tea.Cmd {
	return func() tea.Msg {
		hasMetadata := due != nil || priority > 0 || status != "" || body != ""

		// Plain task with no metadata change → simple description rewrite
		if t.LinkedTaskFile == "" && !hasMetadata {
			if newSummary != t.Description {
				if err := vault.UpdatePlainTaskDescription(t.Source.FilePath, t.Source.Line, newSummary); err != nil {
					return TaskEditedMsg{Task: t, Err: err}
				}
			}
			return TaskEditedMsg{Task: t, NewSummary: newSummary}
		}

		// Plain task with metadata → upgrade to linked task
		if t.LinkedTaskFile == "" {
			taskFilesDir, _ := vault.EnsureTaskFolder(cfg.Vault.Path, cfg.Vault.TaskFilesFolder)
			uid := uuid.New().String()
			var caldavErr error

			if push && cfg.CalDAV.URL != "" {
				pushTask := &task.Task{Description: newSummary, CalDAVUID: uid}
				uid2, err := caldav.PushTask(cfg.CalDAV, pushTask, due, priority, status, body)
				if err != nil {
					caldavErr = err
				} else {
					uid = uid2
				}
			}

			if err := vault.CreateTaskFile(taskFilesDir, uid, newSummary, body, due, priority, status); err != nil {
				return TaskEditedMsg{Task: t, Err: err}
			}
			if err := vault.RewriteTaskLine(t.Source.FilePath, t.Source.Line, uid, newSummary); err != nil {
				return TaskEditedMsg{Task: t, Err: err}
			}

			return TaskEditedMsg{Task: t, NewSummary: newSummary, Reload: true, CalDAVErr: caldavErr}
		}

		// Linked task → update task file
		if newSummary != t.Description {
			if err := vault.UpdateTaskFileTitle(t.LinkedTaskFile, newSummary); err != nil {
				return TaskEditedMsg{Task: t, Err: err}
			}
			uid := filepath.Base(strings.TrimSuffix(t.LinkedTaskFile, ".md"))
			if err := vault.RewriteTaskLine(t.Source.FilePath, t.Source.Line, uid, newSummary); err != nil {
				return TaskEditedMsg{Task: t, Err: err}
			}
		}

		if err := vault.UpdateTaskFileBody(t.LinkedTaskFile, body); err != nil {
			return TaskEditedMsg{Task: t, Err: err}
		}

		if err := vault.UpdateTaskFileFrontmatter(t.LinkedTaskFile, due, status, priority); err != nil {
			return TaskEditedMsg{Task: t, Err: err}
		}

		// Push if already synced to CalDAV (auto-push, no toggle needed)
		var caldavErr error
		if t.CalDAVUID != "" && cfg.CalDAV.URL != "" {
			_, caldavErr = caldav.PushTask(cfg.CalDAV, t, due, priority, status, body)
		} else if push && cfg.CalDAV.URL != "" {
			uid, err := caldav.PushTask(cfg.CalDAV, t, due, priority, status, body)
			if err != nil {
				caldavErr = err
			} else {
				_ = vault.WriteFrontmatterUID(t.LinkedTaskFile, uid)
			}
		}

		return TaskEditedMsg{Task: t, NewSummary: newSummary, Reload: true, CalDAVErr: caldavErr}
	}
}
