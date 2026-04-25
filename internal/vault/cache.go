package vault

import (
	"encoding/json"
	"os"

	"github.com/hawkaii/obia/internal/task"
)

func SaveCache(tasks []task.Task, path string) error {
	data, err := json.Marshal(tasks)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func LoadCache(path string) ([]task.Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tasks []task.Task
	return tasks, json.Unmarshal(data, &tasks)
}
