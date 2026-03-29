package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hawkaii/obia/internal/caldav"
	"github.com/hawkaii/obia/internal/config"
	"github.com/hawkaii/obia/internal/task"
	"github.com/hawkaii/obia/internal/vault"
)

type mode int

const (
	modeNormal mode = iota
	modeFilter
	modeAddTask
)

type App struct {
	cfg       config.Config
	allTasks  []task.Task
	filtered  []task.Task
	activeTab Tab
	cursor    int
	width     int
	height    int
	mode      mode
	input     string
	message   string
	loading   bool
}

func NewApp(cfg config.Config) App {
	return App{
		cfg:       cfg,
		activeTab: TabTasks,
		loading:   true,
	}
}

type tasksLoadedMsg struct {
	tasks []task.Task
}

func loadTasks(vaultPath string) tea.Cmd {
	return func() tea.Msg {
		tasks, err := vault.ParseAllTasks(vaultPath)
		if err != nil {
			return tasksLoadedMsg{}
		}
		return tasksLoadedMsg{tasks: tasks}
	}
}

func (a App) Init() tea.Cmd {
	return loadTasks(a.cfg.Vault.Path)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case tasksLoadedMsg:
		a.allTasks = msg.tasks
		a.loading = false
		a.applyFilter()

	case tea.KeyMsg:
		return a.handleKey(msg)
	}

	return a, nil
}

func (a App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch a.mode {
	case modeFilter:
		return a.handleFilterKey(key, msg)
	case modeAddTask:
		return a.handleAddTaskKey(key, msg)
	default:
		return a.handleNormalKey(key)
	}
}

func (a App) handleNormalKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q", "ctrl+c":
		return a, tea.Quit

	case "j", "down":
		if a.cursor < len(a.filtered)-1 {
			a.cursor++
		}

	case "k", "up":
		if a.cursor > 0 {
			a.cursor--
		}

	case "g":
		a.cursor = 0

	case "G":
		if len(a.filtered) > 0 {
			a.cursor = len(a.filtered) - 1
		}

	case "tab":
		a.activeTab = (a.activeTab + 1) % tabCount
		a.applyFilter()
		a.cursor = 0

	case "shift+tab":
		a.activeTab = (a.activeTab - 1 + tabCount) % tabCount
		a.applyFilter()
		a.cursor = 0

	case "enter":
		if a.cursor < len(a.filtered) {
			t := &a.filtered[a.cursor]
			if err := vault.ToggleTask(t); err != nil {
				a.message = "Error: " + err.Error()
			} else {
				t.Toggle()
				a.syncBack(t)
				a.message = "Toggled task"
			}
			a.applyFilter()
		}

	case "/":
		a.mode = modeFilter
		a.input = ""

	case "a":
		a.mode = modeAddTask
		a.input = ""

	case "p":
		if a.cursor < len(a.filtered) && a.cfg.CalDAV.URL != "" {
			t := &a.filtered[a.cursor]
			uid, err := caldav.PushTask(a.cfg.CalDAV, t)
			if err != nil {
				a.message = "CalDAV push error: " + err.Error()
			} else {
				t.CalDAVUID = uid
				a.syncBack(t)
				a.message = "Pushed to CalDAV: " + t.Description
			}
		} else if a.cfg.CalDAV.URL == "" {
			a.message = "CalDAV not configured"
		}

	case "r":
		a.loading = true
		return a, loadTasks(a.cfg.Vault.Path)
	}

	return a, nil
}

func (a App) handleFilterKey(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		a.applyFilter()
		a.mode = modeNormal
	case "esc":
		a.input = ""
		a.mode = modeNormal
		a.applyFilter()
	case "backspace":
		if len(a.input) > 0 {
			a.input = a.input[:len(a.input)-1]
		}
		a.applyFilter()
	default:
		if len(msg.Runes) > 0 {
			a.input += string(msg.Runes)
			a.applyFilter()
		}
	}
	return a, nil
}

func (a App) handleAddTaskKey(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		if a.input != "" {
			filePath := a.cfg.Vault.Path + "/" + a.cfg.Vault.DefaultTaskFile
			if err := vault.AppendTask(filePath, a.input); err != nil {
				a.message = "Error: " + err.Error()
			} else {
				a.message = "Added: " + a.input
			}
			a.input = ""
			a.mode = modeNormal
			return a, loadTasks(a.cfg.Vault.Path)
		}
		a.mode = modeNormal
	case "esc":
		a.input = ""
		a.mode = modeNormal
	case "backspace":
		if len(a.input) > 0 {
			a.input = a.input[:len(a.input)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			a.input += string(msg.Runes)
		}
	}
	return a, nil
}

func (a *App) applyFilter() {
	tabFiltered := filterTasksForTab(a.allTasks, a.activeTab, a.cfg.Vault.DailyNotesFolder, a.cfg.Vault.DailyNotesFormat)

	if a.input == "" {
		a.filtered = tabFiltered
		return
	}

	query := strings.ToLower(a.input)
	var out []task.Task
	for _, t := range tabFiltered {
		if strings.Contains(strings.ToLower(t.Description), query) {
			out = append(out, t)
		}
	}
	a.filtered = out

	if a.cursor >= len(a.filtered) && len(a.filtered) > 0 {
		a.cursor = len(a.filtered) - 1
	}
}

// syncBack updates the allTasks slice when a filtered task changes.
func (a *App) syncBack(t *task.Task) {
	for i := range a.allTasks {
		if a.allTasks[i].Source == t.Source {
			a.allTasks[i].Status = t.Status
			break
		}
	}
}

func (a App) View() string {
	if a.cfg.Vault.Path == "" {
		return "No vault path configured. Set it in ~/.config/obia/config.toml\n\nPress q to quit."
	}

	w := a.width
	if w < 1 {
		w = 80
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("  Obia"))
	b.WriteString("\n")

	// Tab bar
	b.WriteString(renderTabBar(a.activeTab, w))
	b.WriteString("\n")

	// Task list
	listHeight := a.height - 7 // reserve space for title, tabs, status bar
	if listHeight < 1 {
		listHeight = 10
	}

	if a.loading {
		b.WriteString("  Loading tasks...\n")
	} else if len(a.filtered) == 0 {
		b.WriteString("  No tasks\n")
	} else {
		// Scrolling window
		start := 0
		if a.cursor >= listHeight {
			start = a.cursor - listHeight + 1
		}
		end := start + listHeight
		if end > len(a.filtered) {
			end = len(a.filtered)
		}

		for i := start; i < end; i++ {
			t := a.filtered[i]
			checkbox := "[ ]"
			style := taskTodoStyle
			if t.IsDone() {
				checkbox = "[x]"
				style = taskDoneStyle
			}

			relPath := t.RelativePath(a.cfg.Vault.Path)
			desc := fmt.Sprintf("  %s %s", checkbox, t.Description)
			source := sourceStyle.Render(relPath)

			line := style.Render(desc)
			padding := w - lipgloss.Width(desc) - lipgloss.Width(relPath) - 2
			if padding < 1 {
				padding = 1
			}

			row := line + strings.Repeat(" ", padding) + source
			if i == a.cursor {
				row = selectedStyle.Width(w).Render(row)
			}

			b.WriteString(row)
			b.WriteString("\n")
		}
	}

	// Input line (filter/add mode)
	if a.mode == modeFilter {
		b.WriteString(filterPromptStyle.Render("/") + a.input + "█\n")
	} else if a.mode == modeAddTask {
		b.WriteString(filterPromptStyle.Render("add: ") + a.input + "█\n")
	}

	// Message
	if a.message != "" {
		b.WriteString(messageStyle.Render("  "+a.message) + "\n")
	}

	// Status bar
	bar := "  ↑/k ↓/j navigate  enter: toggle  p: push caldav  /: filter  a: add  tab: switch  r: reload  q: quit"
	b.WriteString(statusBarStyle.Width(w).Render(bar))

	return b.String()
}
