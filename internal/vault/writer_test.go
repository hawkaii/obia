package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestResolveTaskFile_DailyExists(t *testing.T) {
	dir := t.TempDir()
	diary := filepath.Join(dir, "diary")
	os.MkdirAll(diary, 0o755)

	today := time.Now().Format("2006-01-02")
	dailyFile := filepath.Join(diary, today+".md")
	os.WriteFile(dailyFile, []byte("# Today\n"), 0o644)

	got := ResolveTaskFile(dir, "diary", "2006-01-02", "todo.md", "daily")
	if got != dailyFile {
		t.Errorf("expected %s, got %s", dailyFile, got)
	}
}

func TestResolveTaskFile_DailyCreatedFromTemplate(t *testing.T) {
	dir := t.TempDir()
	diary := filepath.Join(dir, "diary")
	os.MkdirAll(diary, 0o755)
	tmplDir := filepath.Join(dir, "templates")
	os.MkdirAll(tmplDir, 0o755)
	os.WriteFile(filepath.Join(tmplDir, "diary template.md"), []byte("**Date**: {{date}}\n"), 0o644)

	today := time.Now().Format("2006-01-02")
	expected := filepath.Join(diary, today+".md")

	got := ResolveTaskFile(dir, "diary", "2006-01-02", "todo.md", "daily")
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}

	data, _ := os.ReadFile(got)
	if !strings.Contains(string(data), today) {
		t.Errorf("template variable not replaced: %q", string(data))
	}
}

func TestResolveTaskFile_DailyCreatedBare(t *testing.T) {
	dir := t.TempDir()
	diary := filepath.Join(dir, "diary")
	os.MkdirAll(diary, 0o755)
	// No templates folder

	today := time.Now().Format("2006-01-02")
	expected := filepath.Join(diary, today+".md")

	got := ResolveTaskFile(dir, "diary", "2006-01-02", "todo.md", "daily")
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}

	data, _ := os.ReadFile(got)
	if !strings.Contains(string(data), "# "+today) {
		t.Errorf("bare heading missing: %q", string(data))
	}
}

func TestResolveTaskFile_DailyFolderMissing(t *testing.T) {
	dir := t.TempDir()
	// No diary folder

	expected := filepath.Join(dir, "todo.md")
	got := ResolveTaskFile(dir, "diary", "2006-01-02", "todo.md", "daily")
	if got != expected {
		t.Errorf("expected fallback %s, got %s", expected, got)
	}
}

func TestResolveTaskFile_DefaultTarget(t *testing.T) {
	dir := t.TempDir()
	expected := filepath.Join(dir, "todo.md")

	got := ResolveTaskFile(dir, "diary", "2006-01-02", "todo.md", "default")
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
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
