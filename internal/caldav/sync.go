package caldav

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/hawkaii/obia/internal/config"
	"github.com/hawkaii/obia/internal/task"
)

// UIDMap tracks which tasks have been pushed to CalDAV.
// Key: "filepath:line" → Value: CalDAV UID
type UIDMap map[string]string

func syncFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "obia", "sync.json"), nil
}

func LoadUIDMap() (UIDMap, error) {
	path, err := syncFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(UIDMap), nil
		}
		return nil, err
	}

	var m UIDMap
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func SaveUIDMap(m UIDMap) error {
	path, err := syncFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func taskKey(t *task.Task) string {
	return fmt.Sprintf("%s:%d", t.Source.FilePath, t.Source.Line)
}

// PushTask pushes a single task to CalDAV, generating a UID if needed.
// description is an optional long-form body separate from the task summary.
func PushTask(cfg config.CalDAV, t *task.Task, start, due *time.Time, rrule string, priority int, status, description string) (string, error) {
	uidMap, err := LoadUIDMap()
	if err != nil {
		return "", fmt.Errorf("loading sync map: %w", err)
	}

	key := taskKey(t)
	uid, exists := uidMap[key]
	if !exists {
		uid = uuid.New().String()
	}
	// Use task's existing UID if it was hydrated from a task file
	if t.CalDAVUID != "" {
		uid = t.CalDAVUID
	}

	icsData := BuildVTodo(uid, t.Description, description, start, due, rrule, priority, status)
	if err := PushTodo(cfg, uid, icsData); err != nil {
		return "", err
	}

	uidMap[key] = uid
	if err := SaveUIDMap(uidMap); err != nil {
		return "", fmt.Errorf("saving sync map: %w", err)
	}

	return uid, nil
}
