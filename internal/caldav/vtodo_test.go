package caldav

import (
	"strings"
	"testing"
	"time"
)

func TestBuildVTodoNoDue(t *testing.T) {
	ics := BuildVTodo("test-uid-123", "Buy groceries", "", nil, nil, "", 0, "")

	if !strings.Contains(ics, "UID:test-uid-123") {
		t.Error("missing UID")
	}
	if !strings.Contains(ics, "SUMMARY:Buy groceries") {
		t.Error("missing SUMMARY")
	}
	if !strings.Contains(ics, "STATUS:NEEDS-ACTION") {
		t.Error("missing STATUS")
	}
	if strings.Contains(ics, "DUE") {
		t.Error("should not contain DUE")
	}
	if strings.Contains(ics, "PRIORITY") {
		t.Error("should not contain PRIORITY when 0")
	}
}

func TestBuildVTodoWithDateDue(t *testing.T) {
	due := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	ics := BuildVTodo("uid-1", "Deploy fix", "", nil, &due, "", 0, "")

	if !strings.Contains(ics, "DUE;VALUE=DATE:20260401") {
		t.Errorf("expected date-only DUE, got: %s", ics)
	}
}

func TestBuildVTodoWithDateTimeDue(t *testing.T) {
	due := time.Date(2026, 4, 1, 14, 30, 0, 0, time.UTC)
	ics := BuildVTodo("uid-2", "Meeting", "", nil, &due, "", 0, "")

	if !strings.Contains(ics, "DUE:20260401T143000Z") {
		t.Errorf("expected datetime DUE, got: %s", ics)
	}
}

func TestBuildVTodoWithPriority(t *testing.T) {
	ics := BuildVTodo("uid-3", "Urgent task", "", nil, nil, "", 5, "")

	if !strings.Contains(ics, "PRIORITY:5") {
		t.Errorf("expected PRIORITY:5, got: %s", ics)
	}
	if !strings.Contains(ics, "STATUS:NEEDS-ACTION") {
		t.Error("expected default STATUS:NEEDS-ACTION")
	}
}

func TestBuildVTodoWithCustomStatus(t *testing.T) {
	ics := BuildVTodo("uid-4", "In progress task", "", nil, nil, "", 0, "IN-PROCESS")

	if !strings.Contains(ics, "STATUS:IN-PROCESS") {
		t.Errorf("expected STATUS:IN-PROCESS, got: %s", ics)
	}
	if strings.Contains(ics, "PRIORITY") {
		t.Error("should not contain PRIORITY when 0")
	}
}

func TestBuildVTodoWithAllOptions(t *testing.T) {
	due := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	ics := BuildVTodo("uid-5", "Full task", "", nil, &due, "", 1, "COMPLETED")

	if !strings.Contains(ics, "STATUS:COMPLETED") {
		t.Errorf("expected STATUS:COMPLETED, got: %s", ics)
	}
	if !strings.Contains(ics, "PRIORITY:1") {
		t.Errorf("expected PRIORITY:1, got: %s", ics)
	}
	if !strings.Contains(ics, "DUE;VALUE=DATE:20260515") {
		t.Errorf("expected DUE date, got: %s", ics)
	}
}

func TestBuildVTodoWithStartAndRRule(t *testing.T) {
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	ics := BuildVTodo("uid-6", "Recurring", "", &start, nil, "FREQ=WEEKLY", 0, "")

	if !strings.Contains(ics, "DTSTART;VALUE=DATE:20260601") {
		t.Errorf("expected date-only DTSTART, got: %s", ics)
	}
	if !strings.Contains(ics, "RRULE:FREQ=WEEKLY") {
		t.Errorf("expected RRULE, got: %s", ics)
	}
}
