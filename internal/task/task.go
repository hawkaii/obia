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
	Tags           []string
	WikiLinks      []string
	Source         Source
	CalDAVUID      string
	LinkedTaskFile string // absolute path to tasks/<uid>.md; empty for plain tasks
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
