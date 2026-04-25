package tasksection

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hawkaii/obia/internal/task"
	"github.com/hawkaii/obia/internal/tui/components/section"
	"github.com/sahilm/fuzzy"
)

// FilterFunc decides which tasks belong in this section.
type FilterFunc func(tasks []task.Task) []task.Task

// Model implements section.Section for a filtered task list view.
type Model struct {
	title     string
	filterFn  FilterFunc
	vaultPath string
	warning   string
	filtered  []task.Task
	search    string
	grouped   bool
}

var _ section.Section = (*Model)(nil)

func New(title, vaultPath string, filterFn FilterFunc) *Model {
	return &Model{
		title:     title,
		filterFn:  filterFn,
		vaultPath: vaultPath,
	}
}

func (m *Model) SetWarning(w string) { m.warning = w }

func (m *Model) Title() string { return m.title }

func (m *Model) SetTasks(all []task.Task) {
	m.filtered = m.filterFn(all)
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
	var content string
	if len(m.filtered) == 0 {
		content = "  No tasks\n"
	} else if m.grouped {
		content = m.viewGrouped(width, height, cursor, selected)
	} else {
		content = m.viewFlat(width, height, cursor, selected)
	}

	if m.warning != "" {
		return warningStyle.Render("  ⚠ "+m.warning) + "\n" + content
	}

	return content
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
		} else if t.CalDAVUID != "" {
			style = taskCalDAVStyle
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
		} else if t.CalDAVUID != "" {
			style = taskCalDAVStyle
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
	taskCalDAVStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00C48C"))
	selectedStyle   = lipgloss.NewStyle().Background(lipgloss.Color("236")).Bold(true)
	sourceStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	fileHeaderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	warningStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// --- Built-in filter functions ---

func FilterOpen(tasks []task.Task) []task.Task {
	return tasks
}

func FilterOverdue(tasks []task.Task) []task.Task {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var out []task.Task
	for i := range tasks {
		if tasks[i].Due != nil && tasks[i].Due.Before(today) {
			out = append(out, tasks[i])
		}
	}
	return out
}

func FilterCalDAV(tasks []task.Task) []task.Task {
	var out []task.Task
	for i := range tasks {
		if tasks[i].CalDAVUID != "" {
			out = append(out, tasks[i])
		}
	}
	return out
}

func MakeFolderFilter(vaultPath string, folders []string) FilterFunc {
	return func(tasks []task.Task) []task.Task {
		var out []task.Task
		for i := range tasks {
			for _, folder := range folders {
				prefix := vaultPath + "/" + folder + "/"
				if strings.HasPrefix(tasks[i].Source.FilePath, prefix) {
					out = append(out, tasks[i])
					break
				}
			}
		}
		return out
	}
}

func MakeTimeWindowFilter(vaultPath string, folders []string, dailyFormat, window, weekStart string) FilterFunc {
	startDay := time.Sunday
	if strings.ToLower(weekStart) == "monday" {
		startDay = time.Monday
	}

	return func(tasks []task.Task) []task.Task {
		now := time.Now()
		var begin, end time.Time

		switch window {
		case "week":
			offset := int(now.Weekday()) - int(startDay)
			if offset < 0 {
				offset += 7
			}
			begin = time.Date(now.Year(), now.Month(), now.Day()-offset, 0, 0, 0, 0, now.Location())
			end = begin.AddDate(0, 0, 7)
		case "month":
			begin = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			end = begin.AddDate(0, 1, 0)
		default:
			return nil
		}

		return matchWindow(tasks, vaultPath, folders, dailyFormat, begin, end)
	}
}

func MakeRollingFilter(vaultPath string, folders []string, dailyFormat string, days int) FilterFunc {
	return func(tasks []task.Task) []task.Task {
		now := time.Now()
		begin := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end := begin.AddDate(0, 0, days)
		return matchWindow(tasks, vaultPath, folders, dailyFormat, begin, end)
	}
}

func matchWindow(tasks []task.Task, vaultPath string, folders []string, dailyFormat string, begin, end time.Time) []task.Task {
	var out []task.Task

outer:
	for i := range tasks {
		t := &tasks[i]

		if t.Due != nil && !t.Due.Before(begin) && t.Due.Before(end) {
			out = append(out, *t)
			continue
		}

		for _, folder := range folders {
			prefix := vaultPath + "/" + folder + "/"
			if !strings.HasPrefix(t.Source.FilePath, prefix) {
				continue
			}

			base := strings.TrimSuffix(filepath.Base(t.Source.FilePath), ".md")
			d, err := time.Parse(dailyFormat, base)
			if err != nil {
				continue
			}

			if !d.Before(begin) && d.Before(end) {
				out = append(out, *t)
				continue outer
			}
		}
	}

	return out
}

func MakeFileFilter(vaultPath, relFile string) FilterFunc {
	abs := filepath.Join(vaultPath, relFile)
	return func(tasks []task.Task) []task.Task {
		var out []task.Task
		for i := range tasks {
			if tasks[i].Source.FilePath == abs {
				out = append(out, tasks[i])
			}
		}
		return out
	}
}

func MakeTagFilter(tag string) FilterFunc {
	tag = strings.TrimPrefix(tag, "#")
	return func(tasks []task.Task) []task.Task {
		var out []task.Task
		for i := range tasks {
			for _, tg := range tasks[i].Tags {
				if tg == tag {
					out = append(out, tasks[i])
					break
				}
			}
		}
		return out
	}
}

func MakeWikiLinkFilter(link string) FilterFunc {
	return func(tasks []task.Task) []task.Task {
		var out []task.Task
		for i := range tasks {
			for _, wl := range tasks[i].WikiLinks {
				if wl == link {
					out = append(out, tasks[i])
					break
				}
			}
		}
		return out
	}
}
