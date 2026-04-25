package tasksection

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hawkaii/obia/internal/task"
	"github.com/hawkaii/obia/internal/tui/components/section"
	"github.com/sahilm/fuzzy"
)

// FilterFunc decides which tasks belong in this section.
type FilterFunc func(tasks []task.Task, folders []string, dailyFormat string) []task.Task

// Model implements section.Section for a filtered task list view.
type Model struct {
	title       string
	filterFn    FilterFunc
	vaultPath   string
	folders     []string
	dailyFormat string
	filtered    []task.Task
	search      string
	grouped     bool
}

var _ section.Section = (*Model)(nil)

func New(title, vaultPath string, folders []string, dailyFormat string, filterFn FilterFunc) *Model {
	return &Model{
		title:       title,
		filterFn:    filterFn,
		vaultPath:   vaultPath,
		folders:     folders,
		dailyFormat: dailyFormat,
	}
}

func (m *Model) Title() string { return m.title }

func (m *Model) SetTasks(all []task.Task) {
	m.filtered = m.filterFn(all, m.folders, m.dailyFormat)
	m.applySearch()
}

func (m *Model) Tasks() []task.Task { return m.filtered }
func (m *Model) NumRows() int       { return len(m.filtered) }

func (m *Model) SetSearch(query string) {
	m.search = query
}

func (m *Model) ToggleGrouped() {
	m.grouped = !m.grouped
}

func (m *Model) IsGrouped() bool {
	return m.grouped
}

func (m *Model) SetGrouped(v bool) {
	m.grouped = v
}

func (m *Model) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	return m, nil
}

func (m *Model) View(width, height, cursor int, selected bool) string {
	if len(m.filtered) == 0 {
		return "  No tasks\n"
	}

	if m.grouped {
		return m.viewGrouped(width, height, cursor, selected)
	}
	return m.viewFlat(width, height, cursor, selected)
}

func (m *Model) viewFlat(width, height, cursor int, selected bool) string {
	var b strings.Builder

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

func (m *Model) viewGrouped(width, height, cursor int, selected bool) string {
	// Build display lines: headers + task rows
	type displayLine struct {
		isHeader  bool
		taskIndex int // -1 for headers
		text      string
	}

	var lines []displayLine
	lastFile := ""
	for i, t := range m.filtered {
		relPath := t.RelativePath(m.vaultPath)
		if relPath != lastFile {
			sep := strings.Repeat("─", width-lipgloss.Width(relPath)-5)
			header := fileHeaderStyle.Render(fmt.Sprintf("── %s %s", relPath, sep))
			lines = append(lines, displayLine{isHeader: true, taskIndex: -1, text: header})
			lastFile = relPath
		}

		checkbox := "[ ]"
		style := taskTodoStyle
		if t.IsDone() {
			checkbox = "[x]"
			style = taskDoneStyle
		}

		row := style.Render(fmt.Sprintf("    %s %s", checkbox, t.Description))
		if selected && i == cursor {
			row = selectedStyle.Width(width).Render(row)
		} else {
			padding := width - lipgloss.Width(row)
			if padding > 0 {
				row = row + strings.Repeat(" ", padding)
			}
		}
		lines = append(lines, displayLine{isHeader: false, taskIndex: i, text: row})
	}

	// Find which display line corresponds to the cursor's task
	cursorDisplayLine := 0
	for i, l := range lines {
		if !l.isHeader && l.taskIndex == cursor {
			cursorDisplayLine = i
			break
		}
	}

	// Scroll window
	start := 0
	if cursorDisplayLine >= height {
		start = cursorDisplayLine - height + 1
	}
	end := start + height
	if end > len(lines) {
		end = len(lines)
	}

	var b strings.Builder
	for i := start; i < end; i++ {
		b.WriteString(lines[i].text)
		b.WriteString("\n")
	}

	return b.String()
}

func (m *Model) applySearch() {
	if m.search == "" {
		return
	}
	source := taskSource(m.filtered)
	matches := fuzzy.FindFrom(m.search, source)
	out := make([]task.Task, 0, len(matches))
	for _, match := range matches {
		out = append(out, m.filtered[match.Index])
	}
	m.filtered = out
}

// taskSource wraps []task.Task to implement fuzzy.Source.
type taskSource []task.Task

func (t taskSource) String(i int) string { return t[i].Description }
func (t taskSource) Len() int            { return len(t) }

// Styles used by the task section renderer.
var (
	taskTodoStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	taskDoneStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Strikethrough(true)
	selectedStyle   = lipgloss.NewStyle().Background(lipgloss.Color("236")).Bold(true)
	sourceStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	fileHeaderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
)

// --- Built-in filter functions ---

func FilterOpen(tasks []task.Task, _ []string, _ string) []task.Task {
	var out []task.Task
	for i := range tasks {
		if !tasks[i].IsDone() {
			out = append(out, tasks[i])
		}
	}
	return out
}

func FilterWeekly(tasks []task.Task, folders []string, dailyFormat string) []task.Task {
	now := time.Now()
	daysBackToSunday := int(now.Weekday()) // 0=Sun, 1=Mon, …, 6=Sat
	weekStart := time.Date(now.Year(), now.Month(), now.Day()-daysBackToSunday, 0, 0, 0, 0, now.Location())
	weekEnd := weekStart.AddDate(0, 0, 7)
	var out []task.Task
outer:
	for i := range tasks {
		t := &tasks[i]
		for _, folder := range folders {
			for d := 0; d < 7; d++ {
				day := weekStart.AddDate(0, 0, d)
				if strings.Contains(t.Source.FilePath, folder+"/"+day.Format(dailyFormat)) {
					if !t.IsDone() {
						out = append(out, *t)
					}
					continue outer
				}
			}
		}
		if t.Due != nil && !t.Due.Before(weekStart) && t.Due.Before(weekEnd) {
			out = append(out, *t)
		}
	}
	return out
}

func FilterDailyFolders(tasks []task.Task, folders []string, _ string) []task.Task {
	var out []task.Task
	for i := range tasks {
		t := &tasks[i]
		for _, folder := range folders {
			if strings.Contains(t.Source.FilePath, "/"+folder+"/") {
				if !t.IsDone() {
					out = append(out, *t)
				}
				break
			}
		}
	}
	return out
}

func FilterOverdue(tasks []task.Task, _ []string, _ string) []task.Task {
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

func FilterCalDAV(tasks []task.Task, _ []string, _ string) []task.Task {
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
