package addform

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
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

	pickerActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Bold(true)

	pickerItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	pickerBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1).
			Width(44)

	ctrlXHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Italic(true)
)

const pickerMaxVisible = 5

// focusIndex constants
const (
	fieldSummary     = 0
	fieldTarget      = 1
	fieldDueDate     = 2
	fieldDueTime     = 3
	fieldDescription = 4
	fieldPriority    = 5
	fieldStatus      = 6
	fieldPush        = 7 // only when showPush
)

// Model is the add task form component.
type Model struct {
	summary     textinput.Model
	targetInput textinput.Model
	targets     []string
	targetSel   int
	pickerIdx   []int
	pickerCur   int
	pickerOff   int
	ctrlXMode   bool
	dueDate     textinput.Model
	dueTime     textinput.Model
	description textinput.Model
	priority    int
	status      int
	push        int
	showPush    bool
	focusIndex  int
	submitted   bool
	cancelled   bool
}

// New creates a new add form. summary pre-fills the Summary field (empty for new tasks).
func New(summary string, targets []string, defaultTargetIdx int, defaultPush bool, showPush bool) Model {
	si := textinput.New()
	si.Placeholder = "Task description"
	si.Prompt = ""
	si.CharLimit = 200
	si.Width = 40
	if summary != "" {
		si.SetValue(summary)
	}
	si.Focus()

	ti := textinput.New()
	ti.Placeholder = "type to filter..."
	ti.Prompt = ""
	ti.CharLimit = 80
	ti.Width = 40

	dd := textinput.New()
	dd.Placeholder = "YYYY-MM-DD"
	dd.Prompt = ""
	dd.CharLimit = 10
	dd.Width = 12
	dd.SetValue(time.Now().AddDate(0, 0, 1).Format("2006-01-02"))

	dt := textinput.New()
	dt.Placeholder = "HH:MM"
	dt.Prompt = ""
	dt.CharLimit = 5
	dt.Width = 7
	dt.SetValue(time.Now().Format("15:04"))

	desc := textinput.New()
	desc.Placeholder = "Optional description"
	desc.Prompt = ""
	desc.CharLimit = 500
	desc.Width = 40

	pushVal := 0
	if defaultPush {
		pushVal = 1
	}

	m := Model{
		summary:     si,
		targetInput: ti,
		targets:     targets,
		targetSel:   defaultTargetIdx,
		dueDate:     dd,
		dueTime:     dt,
		description: desc,
		priority:    0,
		status:      0,
		push:        pushVal,
		showPush:    showPush,
		focusIndex:  fieldSummary,
	}
	m.rebuildPicker()
	return m
}

func (m *Model) numFields() int {
	if m.showPush {
		return 8
	}
	return 7
}

func (m *Model) rebuildPicker() {
	filter := m.targetInput.Value()
	if filter == "" {
		m.pickerIdx = make([]int, len(m.targets))
		for i := range m.targets {
			m.pickerIdx[i] = i
		}
	} else {
		matches := fuzzy.Find(filter, m.targets)
		m.pickerIdx = make([]int, len(matches))
		for i, match := range matches {
			m.pickerIdx[i] = match.Index
		}
	}
	if m.pickerCur >= len(m.pickerIdx) {
		m.pickerCur = max(0, len(m.pickerIdx)-1)
	}
	m.syncPickerOffset()
}

func (m *Model) syncPickerOffset() {
	if m.pickerCur < m.pickerOff {
		m.pickerOff = m.pickerCur
	}
	if m.pickerCur >= m.pickerOff+pickerMaxVisible {
		m.pickerOff = m.pickerCur - pickerMaxVisible + 1
	}
}

func (m *Model) updateFocus() {
	m.summary.Blur()
	m.targetInput.Blur()
	m.dueDate.Blur()
	m.dueTime.Blur()
	m.description.Blur()

	switch m.focusIndex {
	case fieldSummary:
		m.summary.Focus()
	case fieldTarget:
		m.targetInput.Focus()
	case fieldDueDate:
		m.dueDate.Focus()
	case fieldDueTime:
		m.dueTime.Focus()
	case fieldDescription:
		m.description.Focus()
	}
}

func (m *Model) confirmTargetSelection() {
	if len(m.pickerIdx) > 0 && m.pickerCur < len(m.pickerIdx) {
		m.targetSel = m.pickerIdx[m.pickerCur]
	}
}

func (m *Model) selectTargetByName(name string) {
	for i, t := range m.targets {
		if t == name {
			m.targetSel = i
			for j, idx := range m.pickerIdx {
				if idx == i {
					m.pickerCur = j
					m.syncPickerOffset()
					break
				}
			}
			return
		}
	}
}

// isTextField returns true for fields that take text input (no y/n shortcuts).
func (m *Model) isTextField() bool {
	switch m.focusIndex {
	case fieldSummary, fieldTarget, fieldDueDate, fieldDueTime, fieldDescription:
		return true
	}
	return false
}

// Update handles messages for the add form.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := msg.String()

		// ctrl+x chord
		if m.ctrlXMode {
			m.ctrlXMode = false
			switch k {
			case "d":
				m.selectTargetByName("daily")
			case "t":
				m.selectTargetByName("default")
			}
			return m, nil
		}

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
			if m.focusIndex == fieldTarget {
				m.confirmTargetSelection()
			}
			m.focusIndex = (m.focusIndex + 1) % m.numFields()
			m.updateFocus()
			return m, nil

		case "shift+tab":
			if m.focusIndex == fieldTarget {
				m.confirmTargetSelection()
			}
			m.focusIndex = (m.focusIndex - 1 + m.numFields()) % m.numFields()
			m.updateFocus()
			return m, nil

		case "enter":
			if m.focusIndex == fieldSummary {
				if m.summary.Value() != "" {
					m.submitted = true
				}
				return m, nil
			}
			if m.focusIndex == fieldTarget {
				m.confirmTargetSelection()
				m.focusIndex = fieldDueDate
				m.updateFocus()
				return m, nil
			}
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

		// Target picker navigation
		if m.focusIndex == fieldTarget {
			switch k {
			case "ctrl+x":
				m.ctrlXMode = true
				return m, nil
			case "ctrl+n", "down":
				if m.pickerCur < len(m.pickerIdx)-1 {
					m.pickerCur++
					m.syncPickerOffset()
				}
				return m, nil
			case "ctrl+p", "up":
				if m.pickerCur > 0 {
					m.pickerCur--
					m.syncPickerOffset()
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.targetInput, cmd = m.targetInput.Update(msg)
				m.rebuildPicker()
				return m, cmd
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
	case fieldTarget:
		m.targetInput, cmd = m.targetInput.Update(msg)
	case fieldDueDate:
		m.dueDate, cmd = m.dueDate.Update(msg)
	case fieldDueTime:
		m.dueTime, cmd = m.dueTime.Update(msg)
	case fieldDescription:
		m.description, cmd = m.description.Update(msg)
	}
	return m, cmd
}

// View renders the add form.
func (m Model) View() string {
	var b strings.Builder

	b.WriteString(titleFormStyle.Render("Add Task"))
	b.WriteString("\n")

	b.WriteString(m.renderField("Summary", m.summary.View(), m.focusIndex == fieldSummary))
	b.WriteString("\n")

	b.WriteString(m.renderTargetField())
	b.WriteString("\n")

	if m.focusIndex == fieldTarget {
		b.WriteString(m.renderPicker())
		b.WriteString("\n")
	}

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

	hint := "tab/shift+tab: navigate  ←/→: cycle  enter: submit  esc: cancel"
	if m.showPush {
		hint += "  y/n: push"
	}
	b.WriteString(hintStyle.Render(hint))

	return formStyle.Render(b.String())
}

func (m Model) renderTargetField() string {
	label := labelStyle.Render("Target:")
	var value string

	if m.focusIndex == fieldTarget {
		if m.ctrlXMode {
			value = ctrlXHintStyle.Render("^X: d=daily  t=default")
		} else {
			value = m.targetInput.View()
		}
	} else {
		sel := "daily"
		if m.targetSel < len(m.targets) {
			sel = m.targets[m.targetSel]
		}
		value = inactiveFieldStyle.Render("[ " + sel + " ]")
	}

	return fmt.Sprintf("%s %s", label, value)
}

func (m Model) renderPicker() string {
	if len(m.pickerIdx) == 0 {
		return pickerBoxStyle.Render(inactiveFieldStyle.Render("  no matches"))
	}

	var rows []string
	end := m.pickerOff + pickerMaxVisible
	if end > len(m.pickerIdx) {
		end = len(m.pickerIdx)
	}

	for i := m.pickerOff; i < end; i++ {
		label := m.targets[m.pickerIdx[i]]
		if i == m.pickerCur {
			rows = append(rows, pickerActiveStyle.Render("▶ "+label))
		} else {
			rows = append(rows, pickerItemStyle.Render("  "+label))
		}
	}

	content := strings.Join(rows, "\n")
	if m.pickerOff > 0 {
		content = inactiveFieldStyle.Render("  ↑ more") + "\n" + content
	}
	if end < len(m.pickerIdx) {
		content = content + "\n" + inactiveFieldStyle.Render("  ↓ more")
	}

	return pickerBoxStyle.Render(content)
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

func (m Model) GetTarget() string {
	if m.targetSel < len(m.targets) {
		return m.targets[m.targetSel]
	}
	return "daily"
}

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
func (m Model) GetPush() bool     { return m.push == 1 }
