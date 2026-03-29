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

	got, err := ResolveTaskFile(dir, "diary", "2006-01-02", "todo.md", "daily")
	if err != nil {
		t.Fatal(err)
	}
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

	got, err := ResolveTaskFile(dir, "diary", "2006-01-02", "todo.md", "daily")
	if err != nil {
		t.Fatal(err)
	}
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

	got, err := ResolveTaskFile(dir, "diary", "2006-01-02", "todo.md", "daily")
	if err != nil {
		t.Fatal(err)
	}
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
	got, err := ResolveTaskFile(dir, "diary", "2006-01-02", "todo.md", "daily")
	if err != nil {
		t.Fatal(err)
	}
	if got != expected {
		t.Errorf("expected fallback %s, got %s", expected, got)
	}
}

func TestResolveTaskFile_DefaultTarget(t *testing.T) {
	dir := t.TempDir()
	expected := filepath.Join(dir, "todo.md")

	got, err := ResolveTaskFile(dir, "diary", "2006-01-02", "todo.md", "default")
	if err != nil {
		t.Fatal(err)
	}
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

func TestAppendTaskAt_ReturnsCorrectLine(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "todo.md")
	os.WriteFile(f, []byte("# Tasks\n\n- [ ] first\n"), 0o644)

	line, err := AppendTaskAt(f, "second task")
	if err != nil {
		t.Fatal(err)
	}

	// File has 3 lines, AppendTaskAt writes "\n- [ ] second task\n"
	// so the task lands on line 5 (blank line 4, task line 5).
	if line != 5 {
		t.Errorf("expected line 5, got %d", line)
	}

	// Verify the task is actually on that line
	data, _ := os.ReadFile(f)
	lines := strings.Split(string(data), "\n")
	if line-1 >= len(lines) {
		t.Fatalf("line %d out of range, file has %d lines", line, len(lines))
	}
	if !strings.Contains(lines[line-1], "- [ ] second task") {
		t.Errorf("line %d = %q, expected task", line, lines[line-1])
	}
}

func TestAppendTaskAt_NewFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "new.md")

	line, err := AppendTaskAt(f, "first task")
	if err != nil {
		t.Fatal(err)
	}

	if line != 2 {
		t.Errorf("expected line 2, got %d", line)
	}
}
