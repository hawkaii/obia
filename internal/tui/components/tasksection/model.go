package tasksection

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hawkaii/obia/internal/task"
	"github.com/hawkaii/obia/internal/tui/components/section"
)

// FilterFunc decides which tasks belong in this section.
type FilterFunc func(tasks []task.Task, dailyFolder, dailyFormat string) []task.Task

// Model implements section.Section for a filtered task list view.
type Model struct {
	title       string
	filterFn    FilterFunc
	vaultPath   string
	dailyFolder string
	dailyFormat string
	filtered    []task.Task
	search      string
}

var _ section.Section = (*Model)(nil)

func New(title, vaultPath, dailyFolder, dailyFormat string, filterFn FilterFunc) *Model {
	return &Model{
		title:       title,
		filterFn:    filterFn,
		vaultPath:   vaultPath,
		dailyFolder: dailyFolder,
		dailyFormat: dailyFormat,
	}
}

func (m *Model) Title() string { return m.title }

func (m *Model) SetTasks(all []task.Task) {
	m.filtered = m.filterFn(all, m.dailyFolder, m.dailyFormat)
	m.applySearch()
}

func (m *Model) Tasks() []task.Task { return m.filtered }
func (m *Model) NumRows() int       { return len(m.filtered) }

func (m *Model) SetSearch(query string) {
	m.search = query
	// SetTasks must have been called already; re-filter with search
}

func (m *Model) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	return m, nil
}

func (m *Model) View(width, height, cursor int, selected bool) string {
	var b strings.Builder

	if len(m.filtered) == 0 {
		b.WriteString("  No tasks\n")
		return b.String()
	}

	start := 0
	if cursor >= height {
		start = cursor - height + 1
	}
	end := start + height
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := start; i < end; i++ {
		t := m.filtered[i]
		checkbox := "[ ]"
		style := taskTodoStyle
		if t.IsDone() {
			checkbox = "[x]"
			style = taskDoneStyle
		}

		relPath := t.RelativePath(m.vaultPath)
		desc := fmt.Sprintf("  %s %s", checkbox, t.Description)
		source := sourceStyle.Render(relPath)

		line := style.Render(desc)
		padding := width - lipgloss.Width(desc) - lipgloss.Width(relPath) - 2
		if padding < 1 {
			padding = 1
		}

		row := line + strings.Repeat(" ", padding) + source
		if selected && i == cursor {
			row = selectedStyle.Width(width).Render(row)
		}

		b.WriteString(row)
		b.WriteString("\n")
	}

	return b.String()
}

func (m *Model) applySearch() {
	if m.search == "" {
		return
	}
	query := strings.ToLower(m.search)
	var out []task.Task
	for _, t := range m.filtered {
		if strings.Contains(strings.ToLower(t.Description), query) {
			out = append(out, t)
		}
	}
	m.filtered = out
}

// Styles used by the task section renderer.
var (
	taskTodoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	taskDoneStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Strikethrough(true)
	selectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("236")).Bold(true)
	sourceStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// --- Built-in filter functions ---

func FilterOpen(tasks []task.Task, _, _ string) []task.Task {
	var out []task.Task
	for i := range tasks {
		if !tasks[i].IsDone() {
			out = append(out, tasks[i])
		}
	}
	return out
}

func FilterToday(tasks []task.Task, dailyFolder, dailyFormat string) []task.Task {
	today := time.Now()
	todayStr := today.Format(dailyFormat)
	var out []task.Task
	for i := range tasks {
		t := &tasks[i]
		if strings.Contains(t.Source.FilePath, dailyFolder+"/"+todayStr) {
			out = append(out, *t)
			continue
		}
		if t.Due != nil && sameDay(*t.Due, today) {
			out = append(out, *t)
		}
	}
	return out
}

func FilterOverdue(tasks []task.Task, _, _ string) []task.Task {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var out []task.Task
	for i := range tasks {
		t := &tasks[i]
		if t.Due != nil && t.Due.Before(today) && !t.IsDone() {
			out = append(out, *t)
		}
	}
	return out
}

func FilterCalDAV(tasks []task.Task, _, _ string) []task.Task {
	var out []task.Task
	for i := range tasks {
		if tasks[i].CalDAVUID != "" {
			out = append(out, tasks[i])
		}
	}
	return out
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
