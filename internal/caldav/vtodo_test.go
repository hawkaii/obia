package caldav

import (
	"strings"
	"testing"
	"time"
)

func TestBuildVTodoNoDue(t *testing.T) {
	ics := BuildVTodo("test-uid-123", "Buy groceries", nil)

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
}

func TestBuildVTodoWithDateDue(t *testing.T) {
	due := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	ics := BuildVTodo("uid-1", "Deploy fix", &due)

	if !strings.Contains(ics, "DUE;VALUE=DATE:20260401") {
		t.Errorf("expected date-only DUE, got: %s", ics)
	}
}

func TestBuildVTodoWithDateTimeDue(t *testing.T) {
	due := time.Date(2026, 4, 1, 14, 30, 0, 0, time.UTC)
	ics := BuildVTodo("uid-2", "Meeting", &due)

	if !strings.Contains(ics, "DUE:20260401T143000Z") {
		t.Errorf("expected datetime DUE, got: %s", ics)
	}
}
