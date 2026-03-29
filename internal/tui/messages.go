package tui

import "github.com/hawkaii/obia/internal/task"

// TasksLoadedMsg is sent when vault parsing completes.
type TasksLoadedMsg struct {
	Tasks []task.Task
	Err   error
}

// TaskToggledMsg is sent after a task checkbox is written back to disk.
type TaskToggledMsg struct {
	Task *task.Task
	Err  error
}

// TaskAddedMsg is sent after a new task is appended to a file.
type TaskAddedMsg struct {
	Description string
	Err         error
}

// CalDAVPushedMsg is sent after pushing a task to CalDAV.
type CalDAVPushedMsg struct {
	Task *task.Task
	UID  string
	Err  error
}

// ErrorMsg represents a generic error.
type ErrorMsg struct {
	Err error
}
