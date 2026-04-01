package caldav

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/hawkaii/obia/internal/config"
)

// RemoteTodo represents a task fetched from the CalDAV server.
type RemoteTodo struct {
	UID         string
	Status      string
	Summary     string
	Description string
	Due         *time.Time
	Priority    int
}

func basicAuth(username, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}

func collectionURL(cfg config.CalDAV) string {
	return strings.TrimRight(cfg.URL, "/") + "/"
}

// PushTodo uploads a VTODO to the CalDAV server.
func PushTodo(cfg config.CalDAV, uid, icsData string) error {
	url := collectionURL(cfg) + uid + ".ics"

	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(icsData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", basicAuth(cfg.Username, cfg.Password))
	req.Header.Set("Content-Type", "text/calendar; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("CalDAV push failed (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// PullTodos fetches all VTODOs from the CalDAV server.
func PullTodos(cfg config.CalDAV) ([]RemoteTodo, error) {
	url := collectionURL(cfg)

	reportBody := `<?xml version="1.0" encoding="utf-8" ?>` +
		`<C:calendar-query xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav">` +
		`<D:prop><D:getetag/><C:calendar-data/></D:prop>` +
		`<C:filter>` +
		`<C:comp-filter name="VCALENDAR">` +
		`<C:comp-filter name="VTODO"/>` +
		`</C:comp-filter>` +
		`</C:filter>` +
		`</C:calendar-query>`

	req, err := http.NewRequest("REPORT", url, strings.NewReader(reportBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", basicAuth(cfg.Username, cfg.Password))
	req.Header.Set("Content-Type", "application/xml; charset=utf-8")
	req.Header.Set("Depth", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseCalendarResponse(string(body)), nil
}

var (
	calDataRe     = regexp.MustCompile(`(?i)<(?:C:|cal:)?calendar-data[^>]*>([\s\S]*?)</(?:C:|cal:)?calendar-data>`)
	uidRe         = regexp.MustCompile(`(?m)^UID:(.+)$`)
	statusRe      = regexp.MustCompile(`(?m)^STATUS:(.+)$`)
	summaryRe     = regexp.MustCompile(`(?m)^SUMMARY:(.+)$`)
	descriptionRe = regexp.MustCompile(`(?m)^DESCRIPTION:(.+)$`)
	priorityRe    = regexp.MustCompile(`(?m)^PRIORITY:(.+)$`)
	dueRe         = regexp.MustCompile(`(?m)^DUE(?:;VALUE=DATE)?:(.+)$`)
)

func parseCalendarResponse(body string) []RemoteTodo {
	var todos []RemoteTodo

	matches := calDataRe.FindAllStringSubmatch(body, -1)
	for _, m := range matches {
		raw := m[1]
		raw = strings.ReplaceAll(raw, "&lt;", "<")
		raw = strings.ReplaceAll(raw, "&gt;", ">")
		raw = strings.ReplaceAll(raw, "&amp;", "&")

		uidMatch := uidRe.FindStringSubmatch(raw)
		if uidMatch == nil {
			continue
		}

		todo := RemoteTodo{
			UID:    strings.TrimSpace(uidMatch[1]),
			Status: "NEEDS-ACTION",
		}

		if sm := statusRe.FindStringSubmatch(raw); sm != nil {
			todo.Status = strings.TrimSpace(sm[1])
		}
		if sm := summaryRe.FindStringSubmatch(raw); sm != nil {
			todo.Summary = strings.TrimSpace(sm[1])
		}
		if sm := descriptionRe.FindStringSubmatch(raw); sm != nil {
			todo.Description = strings.TrimSpace(sm[1])
		}
		if sm := priorityRe.FindStringSubmatch(raw); sm != nil {
			fmt.Sscanf(strings.TrimSpace(sm[1]), "%d", &todo.Priority)
		}
		if sm := dueRe.FindStringSubmatch(raw); sm != nil {
			todo.Due = parseICalDate(strings.TrimSpace(sm[1]))
		}

		todos = append(todos, todo)
	}

	return todos
}

func parseICalDate(s string) *time.Time {
	// Date-only: 20260301
	if !strings.Contains(s, "T") {
		if t, err := time.Parse("20060102", s); err == nil {
			return &t
		}
		return nil
	}
	// DateTime: 20260301T103000Z
	s = strings.TrimSuffix(s, "Z")
	if t, err := time.Parse("20060102T150405", s); err == nil {
		return &t
	}
	return nil
}

// TestConnection verifies CalDAV server reachability with a PROPFIND.
func TestConnection(cfg config.CalDAV) error {
	url := collectionURL(cfg)

	req, err := http.NewRequest("PROPFIND", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", basicAuth(cfg.Username, cfg.Password))
	req.Header.Set("Depth", "0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}

	return nil
}
