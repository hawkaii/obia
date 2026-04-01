package tui

import "github.com/hawkaii/obia/internal/task"

// TasksLoadedMsg is sent when vault parsing completes.
type TasksLoadedMsg struct {
	Tasks []task.Task
	Err   error
}

// TaskToggledMsg is sent after a task checkbox is written back to disk.
type TaskToggledMsg struct {
	Task      *task.Task
	Err       error
	CalDAVErr error // non-nil if toggle succeeded but CalDAV push failed
}

// TaskAddedMsg is sent after a new task is appended to a file.
type TaskAddedMsg struct {
	Description string
	Err         error
	AutoPushErr error  // non-nil if auto-push failed (task still saved)
	AutoPushUID string // set if auto-push succeeded
}

// CalDAVPushedMsg is sent after pushing a task to CalDAV.
type CalDAVPushedMsg struct {
	Task *task.Task
	UID  string
	Err  error
}

// PullCalDAVMsg is sent after pulling tasks from CalDAV.
type PullCalDAVMsg struct {
	Updated  int    // task files updated
	Created  int    // new task files created from remote
	Notify   string // non-empty if folder conflict occurred
	Err      error
}

// TaskEditedMsg is sent after an edit is saved to disk.
type TaskEditedMsg struct {
	Task       *task.Task
	NewSummary string
	Reload     bool  // true when a plain task was upgraded to a linked task
	Err        error
	CalDAVErr  error // non-nil if save succeeded but CalDAV push failed
}

// ErrorMsg represents a generic error.
type ErrorMsg struct {
	Err error
}
