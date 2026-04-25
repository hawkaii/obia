package caldav

import (
	"fmt"
	"strings"
	"time"
)

// BuildVTodo generates an iCalendar VTODO string.
// priority: 0 = omit, 1-9 per RFC 5545 (1=highest, 9=lowest).
// status: defaults to "NEEDS-ACTION" if empty.
// description: optional long-form body, separate from summary.
func BuildVTodo(uid, summary, description string, start, due *time.Time, rrule string, priority int, status string) string {
	now := formatDateTime(time.Now())

	if status == "" {
		status = "NEEDS-ACTION"
	}

	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\r\n")
	b.WriteString("VERSION:2.0\r\n")
	b.WriteString("PRODID:-//Obia//CLI//EN\r\n")
	b.WriteString("BEGIN:VTODO\r\n")
	fmt.Fprintf(&b, "UID:%s\r\n", uid)
	fmt.Fprintf(&b, "DTSTAMP:%s\r\n", now)
	fmt.Fprintf(&b, "SUMMARY:%s\r\n", summary)
	fmt.Fprintf(&b, "STATUS:%s\r\n", status)

	if description != "" {
		escaped := strings.NewReplacer(`\`, `\\`, ",", `\,`, ";", `\;`, "\n", `\n`).Replace(description)
		fmt.Fprintf(&b, "DESCRIPTION:%s\r\n", escaped)
	}

	if priority > 0 && priority <= 9 {
		fmt.Fprintf(&b, "PRIORITY:%d\r\n", priority)
	}

	if start != nil {
		if start.Hour() == 0 && start.Minute() == 0 && start.Second() == 0 {
			fmt.Fprintf(&b, "DTSTART;VALUE=DATE:%s\r\n", formatDate(*start))
		} else {
			fmt.Fprintf(&b, "DTSTART:%s\r\n", formatDateTime(*start))
		}
	}

	if due != nil {
		if due.Hour() == 0 && due.Minute() == 0 && due.Second() == 0 {
			fmt.Fprintf(&b, "DUE;VALUE=DATE:%s\r\n", formatDate(*due))
		} else {
			fmt.Fprintf(&b, "DUE:%s\r\n", formatDateTime(*due))
		}
	}

	if rrule != "" {
		fmt.Fprintf(&b, "RRULE:%s\r\n", rrule)
	}

	b.WriteString("END:VTODO\r\n")
	b.WriteString("END:VCALENDAR\r\n")
	return b.String()
}

func formatDate(t time.Time) string {
	return t.Format("20060102")
}

func formatDateTime(t time.Time) string {
	return t.Format("20060102T150405Z")
}
