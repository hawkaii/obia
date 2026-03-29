package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hawkaii/obia/internal/task"
)

func TestToggleTask(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.md")
	content := "- [ ] first task\n- [x] second task\n- [ ] third task\n"
	os.WriteFile(f, []byte(content), 0o644)

	// Toggle first task to done
	tk := &task.Task{
		Status: task.Todo,
		Source: task.Source{FilePath: f, Line: 1},
	}
	if err := ToggleTask(tk); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(f)
	lines := strings.Split(string(data), "\n")
	if lines[0] != "- [x] first task" {
		t.Errorf("line 0 = %q", lines[0])
	}

	// Toggle second task to undone
	tk2 := &task.Task{
		Status: task.Done,
		Source: task.Source{FilePath: f, Line: 2},
	}
	if err := ToggleTask(tk2); err != nil {
		t.Fatal(err)
	}

	data, _ = os.ReadFile(f)
	lines = strings.Split(string(data), "\n")
	if lines[1] != "- [ ] second task" {
		t.Errorf("line 1 = %q", lines[1])
	}
}

func TestAppendTask(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "todo.md")
	os.WriteFile(f, []byte("- [ ] existing\n"), 0o644)

	if err := AppendTask(f, "new task"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(f)
	if !strings.Contains(string(data), "- [ ] new task") {
		t.Errorf("file = %q", string(data))
	}
}
