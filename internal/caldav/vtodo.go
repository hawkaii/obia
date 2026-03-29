package caldav

import (
	"fmt"
	"strings"
	"time"
)

// BuildVTodo generates an iCalendar VTODO string.
func BuildVTodo(uid, summary string, due *time.Time) string {
	now := formatDateTime(time.Now())

	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\r\n")
	b.WriteString("VERSION:2.0\r\n")
	b.WriteString("PRODID:-//Obia//CLI//EN\r\n")
	b.WriteString("BEGIN:VTODO\r\n")
	fmt.Fprintf(&b, "UID:%s\r\n", uid)
	fmt.Fprintf(&b, "DTSTAMP:%s\r\n", now)
	fmt.Fprintf(&b, "SUMMARY:%s\r\n", summary)
	b.WriteString("STATUS:NEEDS-ACTION\r\n")

	if due != nil {
		if due.Hour() == 0 && due.Minute() == 0 && due.Second() == 0 {
			fmt.Fprintf(&b, "DUE;VALUE=DATE:%s\r\n", formatDate(*due))
		} else {
			fmt.Fprintf(&b, "DUE:%s\r\n", formatDateTime(*due))
		}
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
