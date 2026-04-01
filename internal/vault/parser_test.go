package vault

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hawkaii/obia/internal/task"
)

func TestParseTasksBasic(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.md")
	content := `# My Tasks

- [ ] buy groceries
- [x] send email
- [ ] check [[project notes]] for updates
- [X] review #urgent PR
- not a task
`
	os.WriteFile(f, []byte(content), 0o644)

	tasks, err := ParseTasks(f)
	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 4 {
		t.Fatalf("expected 4 tasks, got %d", len(tasks))
	}

	// Task 1: open, no links or tags
	if tasks[0].Description != "buy groceries" {
		t.Errorf("task 0 desc = %q", tasks[0].Description)
	}
	if tasks[0].Status != task.Todo {
		t.Error("task 0 should be todo")
	}

	// Task 2: done
	if tasks[1].Status != task.Done {
		t.Error("task 1 should be done")
	}

	// Task 3: wikilink
	if len(tasks[2].WikiLinks) != 1 || tasks[2].WikiLinks[0] != "project notes" {
		t.Errorf("task 2 wikilinks = %v", tasks[2].WikiLinks)
	}

	// Task 4: tag + uppercase X
	if tasks[3].Status != task.Done {
		t.Error("task 3 should be done (uppercase X)")
	}
	if len(tasks[3].Tags) != 1 || tasks[3].Tags[0] != "urgent" {
		t.Errorf("task 3 tags = %v", tasks[3].Tags)
	}
}

func TestParseFrontmatter(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "task.md")
	content := `---
type: task
title: "deploy fix"
due: 2026-04-01
caldav-uid: abc-123
---

- [ ] deploy the fix
`
	os.WriteFile(f, []byte(content), 0o644)

	tasks, err := ParseTasks(f)
	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	if tasks[0].CalDAVUID != "abc-123" {
		t.Errorf("caldav uid = %q", tasks[0].CalDAVUID)
	}
	if tasks[0].Due == nil {
		t.Fatal("due should not be nil")
	}
	if tasks[0].Due.Format("2006-01-02") != "2026-04-01" {
		t.Errorf("due = %v", tasks[0].Due)
	}
}

func TestParseAllTasks(t *testing.T) {
	dir := t.TempDir()

	// Create a couple of files
	os.WriteFile(filepath.Join(dir, "todo.md"), []byte("- [ ] task one\n- [x] task two\n"), 0o644)

	sub := filepath.Join(dir, "diary")
	os.MkdirAll(sub, 0o755)
	os.WriteFile(filepath.Join(sub, "2026-03-29.md"), []byte("- [ ] daily task\n"), 0o644)

	// Skipped directory
	skip := filepath.Join(dir, ".obsidian")
	os.MkdirAll(skip, 0o755)
	os.WriteFile(filepath.Join(skip, "config.md"), []byte("- [ ] should be skipped\n"), 0o644)

	tasks, err := ParseAllTasks(dir, "tasks")
	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}
}
