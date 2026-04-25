package caldav

import "testing"

func TestParseCalendarResponse_StartAndRRule(t *testing.T) {
	body := `<d:multistatus xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav">
<d:response>
<d:propstat><d:prop><c:calendar-data>BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VTODO
UID:abc-123
SUMMARY:Recurring task
STATUS:NEEDS-ACTION
DTSTART;VALUE=DATE:20260601
DUE;VALUE=DATE:20260602
RRULE:FREQ=WEEKLY
PRIORITY:5
END:VTODO
END:VCALENDAR</c:calendar-data></d:prop></d:propstat>
</d:response>
</d:multistatus>`

	todos := parseCalendarResponse(body)
	if len(todos) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(todos))
	}

	td := todos[0]
	if td.Start == nil || td.Start.Format("2006-01-02") != "2026-06-01" {
		t.Fatalf("unexpected start: %v", td.Start)
	}
	if td.RRule != "FREQ=WEEKLY" {
		t.Fatalf("unexpected rrule: %q", td.RRule)
	}
}
