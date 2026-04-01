package editform

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
	pushOptions     = []string{"No", "Yes"}

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

// focusIndex constants
const (
	fieldSummary     = 0
	fieldDueDate     = 1
	fieldDueTime     = 2
	fieldDescription = 3
	fieldPriority    = 4
	fieldStatus      = 5
	fieldPush        = 6 // only when showPush
)

// Model is the edit task form component.
type Model struct {
	summary     textinput.Model
	dueDate     textinput.Model
	dueTime     textinput.Model
	description textinput.Model
	priority    int
	status      int
	push        int
	showPush    bool // true when no CalDAV UID yet and CalDAV is configured
	focusIndex  int
	submitted   bool
	cancelled   bool
}

// New creates an edit form pre-filled from the given task.
// showPush should be true when the task has no CalDAV UID and CalDAV is configured.
func New(t *task.Task, showPush bool) Model {
	si := textinput.New()
	si.Placeholder = "Task description"
	si.Prompt = ""
	si.CharLimit = 200
	si.Width = 40
	si.SetValue(t.Description)
	si.Focus()

	dd := textinput.New()
	dd.Placeholder = "YYYY-MM-DD"
	dd.Prompt = ""
	dd.CharLimit = 10
	dd.Width = 12
	if t.Due != nil {
		dd.SetValue(t.Due.Format("2006-01-02"))
	}

	dt := textinput.New()
	dt.Placeholder = "HH:MM"
	dt.Prompt = ""
	dt.CharLimit = 5
	dt.Width = 7
	if t.Due != nil && (t.Due.Hour() != 0 || t.Due.Minute() != 0) {
		dt.SetValue(t.Due.Format("15:04"))
	}

	desc := textinput.New()
	desc.Placeholder = "Optional description"
	desc.Prompt = ""
	desc.CharLimit = 500
	desc.Width = 40
	desc.SetValue(t.Body)

	// Map task priority to option index
	priorityIdx := 0
	for i, v := range priorityValues {
		if v == t.Priority {
			priorityIdx = i
			break
		}
	}

	// Map task CalDAVStatus to option index
	statusIdx := 0
	for i, s := range statusOptions {
		if s == t.CalDAVStatus {
			statusIdx = i
			break
		}
	}

	return Model{
		summary:     si,
		dueDate:     dd,
		dueTime:     dt,
		description: desc,
		priority:    priorityIdx,
		status:      statusIdx,
		showPush:    showPush,
		focusIndex:  fieldSummary,
	}
}

func (m *Model) numFields() int {
	if m.showPush {
		return 7
	}
	return 6
}

func (m *Model) updateFocus() {
	m.summary.Blur()
	m.dueDate.Blur()
	m.dueTime.Blur()
	m.description.Blur()

	switch m.focusIndex {
	case fieldSummary:
		m.summary.Focus()
	case fieldDueDate:
		m.dueDate.Focus()
	case fieldDueTime:
		m.dueTime.Focus()
	case fieldDescription:
		m.description.Focus()
	}
}

func (m *Model) isTextField() bool {
	switch m.focusIndex {
	case fieldSummary, fieldDueDate, fieldDueTime, fieldDescription:
		return true
	}
	return false
}

// Update handles messages for the edit form.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := msg.String()

		if k == "esc" {
			m.cancelled = true
			return m, nil
		}

		// y/n Push shortcuts when not in a text field
		if m.showPush && !m.isTextField() {
			switch k {
			case "y":
				m.push = 1
				return m, nil
			case "n":
				m.push = 0
				return m, nil
			}
		}

		switch k {
		case "tab":
			m.focusIndex = (m.focusIndex + 1) % m.numFields()
			m.updateFocus()
			return m, nil

		case "shift+tab":
			m.focusIndex = (m.focusIndex - 1 + m.numFields()) % m.numFields()
			m.updateFocus()
			return m, nil

		case "enter":
			if m.summary.Value() != "" {
				m.submitted = true
			}
			return m, nil

		case "left", "right":
			switch m.focusIndex {
			case fieldPriority:
				if k == "right" {
					m.priority = (m.priority + 1) % len(priorityOptions)
				} else {
					m.priority = (m.priority - 1 + len(priorityOptions)) % len(priorityOptions)
				}
				return m, nil
			case fieldStatus:
				if k == "right" {
					m.status = (m.status + 1) % len(statusOptions)
				} else {
					m.status = (m.status - 1 + len(statusOptions)) % len(statusOptions)
				}
				return m, nil
			case fieldPush:
				if k == "right" {
					m.push = (m.push + 1) % len(pushOptions)
				} else {
					m.push = (m.push - 1 + len(pushOptions)) % len(pushOptions)
				}
				return m, nil
			}
		}

		var cmd tea.Cmd
		switch m.focusIndex {
		case fieldSummary:
			m.summary, cmd = m.summary.Update(msg)
		case fieldDueDate:
			m.dueDate, cmd = m.dueDate.Update(msg)
		case fieldDueTime:
			m.dueTime, cmd = m.dueTime.Update(msg)
		case fieldDescription:
			m.description, cmd = m.description.Update(msg)
		}
		return m, cmd
	}

	// Non-key messages (cursor blink, etc.)
	var cmd tea.Cmd
	switch m.focusIndex {
	case fieldSummary:
		m.summary, cmd = m.summary.Update(msg)
	case fieldDueDate:
		m.dueDate, cmd = m.dueDate.Update(msg)
	case fieldDueTime:
		m.dueTime, cmd = m.dueTime.Update(msg)
	case fieldDescription:
		m.description, cmd = m.description.Update(msg)
	}
	return m, cmd
}

// View renders the edit form.
func (m Model) View() string {
	var b strings.Builder

	b.WriteString(titleFormStyle.Render("Edit Task"))
	b.WriteString("\n")

	b.WriteString(m.renderField("Summary", m.summary.View(), m.focusIndex == fieldSummary))
	b.WriteString("\n")

	// Due date + time on one line
	dateActive := m.focusIndex == fieldDueDate
	timeActive := m.focusIndex == fieldDueTime
	dueLabel := labelStyle.Render("Due:")
	dateView := m.dueDate.View()
	timeView := m.dueTime.View()
	if !dateActive {
		dateView = inactiveFieldStyle.Render(m.dueDate.Value())
	}
	if !timeActive {
		timeView = inactiveFieldStyle.Render(m.dueTime.Value())
	}
	b.WriteString(fmt.Sprintf("%s %s  %s", dueLabel, dateView, timeView))
	b.WriteString("\n")

	b.WriteString(m.renderField("Desc", m.description.View(), m.focusIndex == fieldDescription))
	b.WriteString("\n")

	b.WriteString(m.renderField("Priority", m.renderCycle(priorityOptions, m.priority, m.focusIndex == fieldPriority), m.focusIndex == fieldPriority))
	b.WriteString("\n")

	b.WriteString(m.renderField("Status", m.renderCycle(statusOptions, m.status, m.focusIndex == fieldStatus), m.focusIndex == fieldStatus))
	b.WriteString("\n")

	if m.showPush {
		b.WriteString(m.renderField("Push?", m.renderCycle(pushOptions, m.push, m.focusIndex == fieldPush), m.focusIndex == fieldPush))
		b.WriteString("\n")
	}

	hint := "tab/shift+tab: navigate  ←/→: cycle  enter: save  esc: cancel"
	if m.showPush {
		hint += "  y/n: push"
	}
	b.WriteString(hintStyle.Render(hint))

	return formStyle.Render(b.String())
}

func (m Model) renderField(label, value string, _ bool) string {
	return fmt.Sprintf("%s %s", labelStyle.Render(label+":"), value)
}

func (m Model) renderCycle(options []string, selected int, active bool) string {
	opt := options[selected]
	if active {
		return fmt.Sprintf("<< %s >>", activeFieldStyle.Render("[ "+opt+" ]"))
	}
	return inactiveFieldStyle.Render("[ " + opt + " ]")
}

func (m Model) Submitted() bool { return m.submitted }
func (m Model) Cancelled() bool { return m.cancelled }

func (m Model) GetSummary() string     { return strings.TrimSpace(m.summary.Value()) }
func (m Model) GetDescription() string { return strings.TrimSpace(m.description.Value()) }
func (m Model) GetPush() bool          { return m.push == 1 }

// GetDue combines the date and time fields into a *time.Time.
// Returns nil if the date field is empty or unparseable.
func (m Model) GetDue() *time.Time {
	dateVal := strings.TrimSpace(m.dueDate.Value())
	if dateVal == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", dateVal)
	if err != nil {
		return nil
	}
	timeVal := strings.TrimSpace(m.dueTime.Value())
	if timeVal != "" {
		var h, min int
		if _, err := fmt.Sscanf(timeVal, "%d:%d", &h, &min); err == nil {
			t = time.Date(t.Year(), t.Month(), t.Day(), h, min, 0, 0, t.Location())
		}
	}
	return &t
}

func (m Model) GetPriority() int  { return priorityValues[m.priority] }
func (m Model) GetStatus() string { return statusOptions[m.status] }
