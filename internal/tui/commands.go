package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hawkaii/obia/internal/caldav"
	"github.com/hawkaii/obia/internal/config"
	"github.com/hawkaii/obia/internal/task"
	"github.com/hawkaii/obia/internal/vault"
)

// LoadTasksCmd scans the vault and returns all parsed tasks.
func LoadTasksCmd(vaultPath string) tea.Cmd {
	return func() tea.Msg {
		tasks, err := vault.ParseAllTasks(vaultPath)
		return TasksLoadedMsg{Tasks: tasks, Err: err}
	}
}

// ToggleTaskCmd writes the toggled checkbox back to the source file.
func ToggleTaskCmd(t *task.Task) tea.Cmd {
	return func() tea.Msg {
		err := vault.ToggleTask(t)
		return TaskToggledMsg{Task: t, Err: err}
	}
}

// AddTaskCmd appends a new task to the given file.
func AddTaskCmd(filePath, description string) tea.Cmd {
	return func() tea.Msg {
		err := vault.AppendTask(filePath, description)
		return TaskAddedMsg{Description: description, Err: err}
	}
}

// PushCalDAVCmd pushes a task to the CalDAV server.
func PushCalDAVCmd(cfg config.CalDAV, t *task.Task) tea.Cmd {
	return func() tea.Msg {
		uid, err := caldav.PushTask(cfg, t, nil, 0, "")
		return CalDAVPushedMsg{Task: t, UID: uid, Err: err}
	}
}

// AddTaskWithMetaCmd appends a task and optionally pushes to CalDAV with full metadata.
// push=true performs the CalDAV push; push=false skips it regardless of caldavCfg.AutoPush.
func AddTaskWithMetaCmd(filePath, description string, due *time.Time, priority int, status string, push bool, caldavCfg config.CalDAV) tea.Cmd {
	return func() tea.Msg {
		line, err := vault.AppendTaskAt(filePath, description)
		if err != nil {
			return TaskAddedMsg{Description: description, Err: err}
		}

		if push && caldavCfg.URL != "" {
			t := &task.Task{
				Description: description,
				Source: task.Source{
					FilePath: filePath,
					Line:     line,
				},
			}
			uid, pushErr := caldav.PushTask(caldavCfg, t, due, priority, status)
			if pushErr != nil {
				return TaskAddedMsg{
					Description: description,
					AutoPushErr: pushErr,
				}
			}
			_ = vault.WriteFrontmatterUID(filePath, uid)
			return TaskAddedMsg{
				Description: description,
				AutoPushUID: uid,
			}
		}

		return TaskAddedMsg{Description: description}
	}
}
