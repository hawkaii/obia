package pushform

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hawkaii/obia/internal/task"
)

var (
	priorityOptions = []string{"none", "1 (highest)", "5 (normal)", "9 (lowest)"}
	priorityValues  = []int{0, 1, 5, 9}
	statusOptions   = []string{"NEEDS-ACTION", "IN-PROCESS", "COMPLETED", "CANCELLED"}

	formStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("170")).
			Padding(1, 2).
			Width(60)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true).
			Width(12)

	activeFieldStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170"))

	inactiveFieldStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245"))

	titleFormStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			MarginBottom(1)

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
)

// Model is the push form component.
type Model struct {
	task       *task.Task
	summary    textinput.Model
	due        textinput.Model
	priority   int  // index into priorityOptions
	status     int  // index into statusOptions
	focusIndex int  // which field is focused (0-3)
	submitted  bool
	cancelled  bool
}

// New creates a new push form pre-filled with task data.
func New(t *task.Task) Model {
	summary := textinput.New()
	summary.Placeholder = "Task summary"
	summary.SetValue(t.Description)
	summary.Focus()
	summary.CharLimit = 200
	summary.Width = 40

	due := textinput.New()
	due.Placeholder = "YYYY-MM-DD"
	due.CharLimit = 10
	due.Width = 40
	if t.Due != nil {
		due.SetValue(t.Due.Format("2006-01-02"))
	}

	return Model{
		task:       t,
		summary:    summary,
		due:        due,
		priority:   0, // "none"
		status:     0, // "NEEDS-ACTION"
		focusIndex: 0,
	}
}

// Update handles messages for the push form.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		switch key {
		case "esc":
			m.cancelled = true
			return m, nil

		case "enter":
			m.submitted = true
			return m, nil

		case "tab":
			m.focusIndex = (m.focusIndex + 1) % 4
			m.updateFocus()
			return m, nil

		case "shift+tab":
			m.focusIndex = (m.focusIndex - 1 + 4) % 4
			m.updateFocus()
			return m, nil

		case "left", "right":
			if m.focusIndex == 2 {
				// priority field
				if key == "right" {
					m.priority = (m.priority + 1) % len(priorityOptions)
				} else {
					m.priority = (m.priority - 1 + len(priorityOptions)) % len(priorityOptions)
				}
				return m, nil
			}
			if m.focusIndex == 3 {
				// status field
				if key == "right" {
					m.status = (m.status + 1) % len(statusOptions)
				} else {
					m.status = (m.status - 1 + len(statusOptions)) % len(statusOptions)
				}
				return m, nil
			}
		}

		// Delegate to focused text input
		var cmd tea.Cmd
		switch m.focusIndex {
		case 0:
			m.summary, cmd = m.summary.Update(msg)
		case 1:
			m.due, cmd = m.due.Update(msg)
		}
		return m, cmd
	}

	// Handle non-key messages (e.g. blink)
	var cmd tea.Cmd
	switch m.focusIndex {
	case 0:
		m.summary, cmd = m.summary.Update(msg)
	case 1:
		m.due, cmd = m.due.Update(msg)
	}
	return m, cmd
}

func (m *Model) updateFocus() {
	m.summary.Blur()
	m.due.Blur()

	switch m.focusIndex {
	case 0:
		m.summary.Focus()
	case 1:
		m.due.Focus()
	}
}

// View renders the push form.
func (m Model) View() string {
	var b strings.Builder

	b.WriteString(titleFormStyle.Render("Push to CalDAV"))
	b.WriteString("\n")

	// Summary
	b.WriteString(m.renderField(0, "Summary", m.summary.View()))
	b.WriteString("\n")

	// Due
	b.WriteString(m.renderField(1, "Due", m.due.View()))
	b.WriteString("\n")

	// Priority
	b.WriteString(m.renderField(2, "Priority", m.renderCycle(priorityOptions, m.priority, m.focusIndex == 2)))
	b.WriteString("\n")

	// Status
	b.WriteString(m.renderField(3, "Status", m.renderCycle(statusOptions, m.status, m.focusIndex == 3)))
	b.WriteString("\n")

	b.WriteString(hintStyle.Render("tab/shift+tab: navigate  left/right: cycle  enter: submit  esc: cancel"))

	return formStyle.Render(b.String())
}

func (m Model) renderField(idx int, label, value string) string {
	l := labelStyle.Render(label + ":")
	if m.focusIndex == idx {
		return fmt.Sprintf("%s %s", l, value)
	}
	return fmt.Sprintf("%s %s", l, value)
}

func (m Model) renderCycle(options []string, selected int, active bool) string {
	var parts []string
	for i, opt := range options {
		if i == selected {
			if active {
				parts = append(parts, activeFieldStyle.Render("[ "+opt+" ]"))
			} else {
				parts = append(parts, inactiveFieldStyle.Render("[ "+opt+" ]"))
			}
		}
	}
	if active {
		return fmt.Sprintf("<< %s >>", strings.Join(parts, " "))
	}
	return strings.Join(parts, " ")
}

// Submitted returns true if the user pressed enter.
func (m Model) Submitted() bool { return m.submitted }

// Cancelled returns true if the user pressed escape.
func (m Model) Cancelled() bool { return m.cancelled }

// GetSummary returns the summary text.
func (m Model) GetSummary() string { return m.summary.Value() }

// GetDue returns the parsed due date, or nil if empty/invalid.
func (m Model) GetDue() *time.Time {
	val := strings.TrimSpace(m.due.Value())
	if val == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", val)
	if err != nil {
		return nil
	}
	return &t
}

// GetPriority returns the selected priority value (0 = none).
func (m Model) GetPriority() int { return priorityValues[m.priority] }

// GetStatus returns the selected CalDAV status string.
func (m Model) GetStatus() string { return statusOptions[m.status] }
