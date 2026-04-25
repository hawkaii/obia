package task

import "time"

type Status int

const (
	Todo Status = iota
	Done
)

func (s Status) String() string {
	switch s {
	case Done:
		return "x"
	default:
		return " "
	}
}

type Source struct {
	FilePath string    // Absolute path to .md file
	Line     int       // Line number (1-indexed)
	FileMod  time.Time // File modification time (for sorting by recency)
}

type Task struct {
	Description    string
	Status         Status
	Due            *time.Time
	Start          *time.Time
	RRule          string
	Tags           []string
	WikiLinks      []string
	Source         Source
	CalDAVUID      string
	LinkedTaskFile string // absolute path to tasks/<uid>.md; empty for plain tasks
	Priority       int    // 0=none, 1-9 per RFC 5545 (linked tasks only)
	CalDAVStatus   string // "NEEDS-ACTION", "IN-PROCESS", etc (linked tasks only)
	Body           string // description body from task file (linked tasks only)
}

func (t *Task) Toggle() {
	if t.Status == Todo {
		t.Status = Done
	} else {
		t.Status = Todo
	}
}

func (t *Task) IsDone() bool {
	return t.Status == Done
}

func (t *Task) RelativePath(vaultPath string) string {
	if len(t.Source.FilePath) > len(vaultPath) {
		return t.Source.FilePath[len(vaultPath)+1:]
	}
	return t.Source.FilePath
}
