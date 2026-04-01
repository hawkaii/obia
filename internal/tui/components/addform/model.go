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

// Model is the add task form component.
type Model struct {
	summary     textinput.Model
	targetInput textinput.Model // fuzzy filter for target picker
	targets     []string        // ["daily", "default", ...extra]
	targetSel   int             // confirmed selection index
	pickerIdx   []int           // filtered indices into targets
	pickerCur   int             // cursor within pickerIdx
	pickerOff   int             // scroll offset
	ctrlXMode   bool            // waiting for chord second key
	due         textinput.Model
	priority    int  // index into priorityOptions
	status      int  // index into statusOptions
	push        int  // 0=No, 1=Yes
	showPush    bool // false when CalDAV URL is empty
	focusIndex  int  // 0=summary 1=target 2=due 3=priority 4=status [5=push]
	submitted   bool
	cancelled   bool
}

// New creates a new add form.
func New(targets []string, defaultTargetIdx int, defaultPush bool, showPush bool) Model {
	summary := textinput.New()
	summary.Placeholder = "Task description"
	summary.Focus()
	summary.CharLimit = 200
	summary.Width = 40

	targetInput := textinput.New()
	targetInput.Placeholder = "type to filter..."
	targetInput.CharLimit = 80
	targetInput.Width = 40

	due := textinput.New()
	due.Placeholder = "YYYY-MM-DD"
	due.CharLimit = 10
	due.Width = 40
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	due.SetValue(tomorrow)

	pushVal := 0
	if defaultPush {
		pushVal = 1
	}

	m := Model{
		summary:     summary,
		targetInput: targetInput,
		targets:     targets,
		targetSel:   defaultTargetIdx,
		due:         due,
		priority:    0,
		status:      0,
		push:        pushVal,
		showPush:    showPush,
		focusIndex:  0,
	}
	m.rebuildPicker()
	return m
}

// rebuildPicker refilters targets based on current targetInput value.
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

func (m *Model) numFields() int {
	if m.showPush {
		return 6
	}
	return 5
}

func (m *Model) updateFocus() {
	m.summary.Blur()
	m.targetInput.Blur()
	m.due.Blur()

	switch m.focusIndex {
	case 0:
		m.summary.Focus()
	case 1:
		m.targetInput.Focus()
	case 2:
		m.due.Focus()
	}
}

// confirmTargetSelection locks in the currently highlighted picker item.
func (m *Model) confirmTargetSelection() {
	if len(m.pickerIdx) > 0 && m.pickerCur < len(m.pickerIdx) {
		m.targetSel = m.pickerIdx[m.pickerCur]
	}
}

// selectTargetByName finds the first target matching name and selects it.
func (m *Model) selectTargetByName(name string) {
	for i, t := range m.targets {
		if t == name {
			m.targetSel = i
			// Reset picker cursor to this item
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

// Update handles messages for the add form.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := msg.String()

		// Handle ctrl+x chord mode
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

		// Global shortcuts
		switch k {
		case "esc":
			m.cancelled = true
			return m, nil
		}

		// y/n Push shortcuts when not in a text input field
		if m.showPush && m.focusIndex != 0 && m.focusIndex != 1 && m.focusIndex != 2 {
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
			if m.focusIndex == 1 {
				m.confirmTargetSelection()
			}
			m.focusIndex = (m.focusIndex + 1) % m.numFields()
			m.updateFocus()
			return m, nil

		case "shift+tab":
			if m.focusIndex == 1 {
				m.confirmTargetSelection()
			}
			m.focusIndex = (m.focusIndex - 1 + m.numFields()) % m.numFields()
			m.updateFocus()
			return m, nil

		case "enter":
			if m.focusIndex == 0 {
				// Fast path: submit immediately from summary
				if m.summary.Value() != "" {
					m.submitted = true
				}
				return m, nil
			}
			if m.focusIndex == 1 {
				// Confirm target and advance to Due
				m.confirmTargetSelection()
				m.focusIndex = 2
				m.updateFocus()
				return m, nil
			}
			// All other fields: submit
			if m.summary.Value() != "" {
				m.submitted = true
			}
			return m, nil

		case "left", "right":
			switch m.focusIndex {
			case 3:
				if k == "right" {
					m.priority = (m.priority + 1) % len(priorityOptions)
				} else {
					m.priority = (m.priority - 1 + len(priorityOptions)) % len(priorityOptions)
				}
				return m, nil
			case 4:
				if k == "right" {
					m.status = (m.status + 1) % len(statusOptions)
				} else {
					m.status = (m.status - 1 + len(statusOptions)) % len(statusOptions)
				}
				return m, nil
			case 5:
				if k == "right" {
					m.push = (m.push + 1) % len(pushOptions)
				} else {
					m.push = (m.push - 1 + len(pushOptions)) % len(pushOptions)
				}
				return m, nil
			}
		}

		// Target picker navigation
		if m.focusIndex == 1 {
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
				// Delegate typing to targetInput
				var cmd tea.Cmd
				m.targetInput, cmd = m.targetInput.Update(msg)
				m.rebuildPicker()
				return m, cmd
			}
		}

		// Delegate text input
		var cmd tea.Cmd
		switch m.focusIndex {
		case 0:
			m.summary, cmd = m.summary.Update(msg)
		case 2:
			m.due, cmd = m.due.Update(msg)
		}
		return m, cmd
	}

	// Non-key messages (cursor blink, etc.)
	var cmd tea.Cmd
	switch m.focusIndex {
	case 0:
		m.summary, cmd = m.summary.Update(msg)
	case 1:
		m.targetInput, cmd = m.targetInput.Update(msg)
	case 2:
		m.due, cmd = m.due.Update(msg)
	}
	return m, cmd
}

// View renders the add form.
func (m Model) View() string {
	var b strings.Builder

	b.WriteString(titleFormStyle.Render("Add Task"))
	b.WriteString("\n")

	// Summary
	b.WriteString(m.renderField("Summary", m.summary.View(), m.focusIndex == 0))
	b.WriteString("\n")

	// Target
	targetDisplay := m.renderTargetField()
	b.WriteString(targetDisplay)
	b.WriteString("\n")

	// Picker (only when Target is focused)
	if m.focusIndex == 1 {
		b.WriteString(m.renderPicker())
		b.WriteString("\n")
	}

	// Due
	b.WriteString(m.renderField("Due", m.due.View(), m.focusIndex == 2))
	b.WriteString("\n")

	// Priority
	b.WriteString(m.renderField("Priority", m.renderCycle(priorityOptions, m.priority, m.focusIndex == 3), m.focusIndex == 3))
	b.WriteString("\n")

	// Status
	b.WriteString(m.renderField("Status", m.renderCycle(statusOptions, m.status, m.focusIndex == 4), m.focusIndex == 4))
	b.WriteString("\n")

	// Push (optional)
	if m.showPush {
		b.WriteString(m.renderField("Push?", m.renderCycle(pushOptions, m.push, m.focusIndex == 5), m.focusIndex == 5))
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

	if m.focusIndex == 1 {
		// Show filter input and ctrl+x hint
		if m.ctrlXMode {
			value = ctrlXHintStyle.Render("^X: d=daily  t=default")
		} else {
			value = m.targetInput.View()
		}
	} else {
		// Show confirmed selection
		sel := "daily"
		if m.targetSel < len(m.targets) {
			sel = m.targets[m.targetSel]
		}
		value = activeFieldStyle.Render("[ " + sel + " ]")
		if m.focusIndex != 1 {
			value = inactiveFieldStyle.Render("[ " + sel + " ]")
		}
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

	// Scroll indicators
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
	l := labelStyle.Render(label + ":")
	return fmt.Sprintf("%s %s", l, value)
}

func (m Model) renderCycle(options []string, selected int, active bool) string {
	opt := options[selected]
	if active {
		return fmt.Sprintf("<< %s >>", activeFieldStyle.Render("[ "+opt+" ]"))
	}
	return inactiveFieldStyle.Render("[ " + opt + " ]")
}

// Submitted returns true if the user submitted the form.
func (m Model) Submitted() bool { return m.submitted }

// Cancelled returns true if the user pressed escape.
func (m Model) Cancelled() bool { return m.cancelled }

// GetSummary returns the task summary.
func (m Model) GetSummary() string { return m.summary.Value() }

// GetTarget returns the selected target string (e.g. "daily", "default", "projects/work.md").
func (m Model) GetTarget() string {
	if m.targetSel < len(m.targets) {
		return m.targets[m.targetSel]
	}
	return "daily"
}

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

// GetPush returns true if the user chose to push to CalDAV.
func (m Model) GetPush() bool { return m.push == 1 }
