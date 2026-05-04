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

func TestAppendTaskAt_NoTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "note.md")
	os.WriteFile(f, []byte("foo"), 0o644) // no trailing newline

	line, err := AppendTaskAt(f, "my task")
	if err != nil {
		t.Fatal(err)
	}

	if line != 2 {
		t.Errorf("expected line 2, got %d", line)
	}

	data, _ := os.ReadFile(f)
	lines := strings.Split(string(data), "\n")
	if line-1 >= len(lines) {
		t.Fatalf("line %d out of range, file has %d lines", line, len(lines))
	}
	if !strings.Contains(lines[line-1], "- [ ] my task") {
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

func TestWriteFrontmatterUID_SingleTaskFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "task.md")
	content := "---\ntype: task\ndue: 2026-04-01\n---\n\n- [ ] Do the thing\n"
	os.WriteFile(f, []byte(content), 0o644)

	if err := WriteFrontmatterUID(f, "abc-123"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(f)
	if !strings.Contains(string(data), "caldav-uid: abc-123") {
		t.Errorf("expected caldav-uid in file, got:\n%s", string(data))
	}
}

func TestWriteFrontmatterUID_UpdatesExisting(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "task.md")
	content := "---\ntype: task\ncaldav-uid: old-uid\n---\n\n- [ ] Do the thing\n"
	os.WriteFile(f, []byte(content), 0o644)

	if err := WriteFrontmatterUID(f, "new-uid"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(f)
	result := string(data)
	if !strings.Contains(result, "caldav-uid: new-uid") {
		t.Errorf("expected updated caldav-uid, got:\n%s", result)
	}
	if strings.Contains(result, "old-uid") {
		t.Errorf("expected old-uid to be replaced, got:\n%s", result)
	}
}

func TestWriteFrontmatterUID_SkipsMultiTaskFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "todo.md")
	content := "---\ntitle: My Todo List\n---\n\n- [ ] Task one\n- [ ] Task two\n"
	os.WriteFile(f, []byte(content), 0o644)

	if err := WriteFrontmatterUID(f, "some-uid"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(f)
	if strings.Contains(string(data), "caldav-uid") {
		t.Errorf("expected no caldav-uid written to multi-task file, got:\n%s", string(data))
	}
}

func TestWriteFrontmatterUID_SkipsNoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "plain.md")
	content := "# Plain Note\n\n- [ ] Task without frontmatter\n"
	os.WriteFile(f, []byte(content), 0o644)

	if err := WriteFrontmatterUID(f, "some-uid"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(f)
	if strings.Contains(string(data), "caldav-uid") {
		t.Errorf("expected no caldav-uid written to file without frontmatter, got:\n%s", string(data))
	}
}

func TestCreateTaskFile_WritesStartAndRRule(t *testing.T) {
	dir := t.TempDir()
	start := time.Date(2026, 3, 30, 9, 0, 0, 0, time.UTC)
	due := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)

	if err := CreateTaskFile(dir, "uid-1", "Title", "Body", &start, &due, "FREQ=WEEKLY", 5, "NEEDS-ACTION"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "uid-1.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, "dtstart: 2026-03-30T09:00:00Z") {
		t.Fatalf("missing dtstart in file:\n%s", s)
	}
	if !strings.Contains(s, "rrule: FREQ=WEEKLY") {
		t.Fatalf("missing rrule in file:\n%s", s)
	}
}
