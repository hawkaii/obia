package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hawkaii/obia/internal/caldav"
	"github.com/hawkaii/obia/internal/config"
	"github.com/hawkaii/obia/internal/task"
	"github.com/hawkaii/obia/internal/tui/components/addform"
	"github.com/hawkaii/obia/internal/tui/components/pushform"
	"github.com/hawkaii/obia/internal/vault"
	appctx "github.com/hawkaii/obia/internal/tui/context"
	"github.com/hawkaii/obia/internal/tui/keys"
	"github.com/hawkaii/obia/internal/tui/components/section"
	"github.com/hawkaii/obia/internal/tui/components/tasksection"
)

// appMode controls which top-level mode the TUI is in.
type appMode int

const (
	modeBrowser appMode = iota
	modePushForm
	modeAddForm
	// modeChat — future
)

// inputMode controls what the text input is doing.
type inputMode int

const (
	inputNone inputMode = iota
	inputFilter
)

type App struct {
	ctx       *appctx.ProgramContext
	keys      keys.KeyMap
	mode      appMode
	inputMode inputMode
	input     string
	message   string

	// Browser state
	allTasks  []task.Task
	sections  []section.Section
	activeTab int
	cursor    int
	loading   bool
	pushForm  pushform.Model
	addForm   addform.Model
}

func NewApp(cfg config.Config) App {
	ctx := appctx.New(cfg)

	vp := cfg.Vault.Path
	df := cfg.Vault.DailyNotesFolder
	dfmt := cfg.Vault.DailyNotesFormat

	sections := []section.Section{
		tasksection.New("Tasks", vp, df, dfmt, tasksection.FilterOpen),
		tasksection.New("Today", vp, df, dfmt, tasksection.FilterToday),
		tasksection.New("Overdue", vp, df, dfmt, tasksection.FilterOverdue),
		tasksection.New("CalDAV", vp, df, dfmt, tasksection.FilterCalDAV),
	}

	return App{
		ctx:      ctx,
		keys:     keys.DefaultKeyMap,
		mode:     modeBrowser,
		sections: sections,
		loading:  true,
	}
}

func (a App) Init() tea.Cmd {
	return LoadTasksCmd(a.ctx.VaultPath())
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.ctx.SetSize(msg.Width, msg.Height)

	case TasksLoadedMsg:
		if msg.Err != nil {
			a.message = "Error loading tasks: " + msg.Err.Error()
		} else {
			a.allTasks = msg.Tasks
			a.refreshSections()
		}
		a.loading = false

	case TaskToggledMsg:
		if msg.Err != nil {
			a.message = "Error: " + msg.Err.Error()
		} else {
			msg.Task.Toggle()
			a.syncBack(msg.Task)
			a.refreshSections()
			a.message = "Toggled task"
		}

	case TaskAddedMsg:
		if msg.Err != nil {
			a.message = "Error: " + msg.Err.Error()
		} else if msg.AutoPushErr != nil {
			a.message = "Task added · CalDAV push failed: " + msg.AutoPushErr.Error()
		} else if msg.AutoPushUID != "" {
			a.message = "Task added · pushed to CalDAV"
		} else {
			a.message = "Task added"
		}
		return a, LoadTasksCmd(a.ctx.VaultPath())

	case CalDAVPushedMsg:
		if msg.Err != nil {
			a.message = "CalDAV push error: " + msg.Err.Error()
		} else {
			msg.Task.CalDAVUID = msg.UID
			a.syncBack(msg.Task)
			// Persist UID to frontmatter for single-task files
			_ = vault.WriteFrontmatterUID(msg.Task.Source.FilePath, msg.UID)
			a.refreshSections()
			a.message = "Pushed to CalDAV: " + msg.Task.Description
		}

	case tea.KeyMsg:
		return a.handleKey(msg)
	}

	// Delegate non-key messages to active form for cursor blink etc.
	if a.mode == modePushForm {
		var cmd tea.Cmd
		a.pushForm, cmd = a.pushForm.Update(msg)
		return a, cmd
	}
	if a.mode == modeAddForm {
		var cmd tea.Cmd
		a.addForm, cmd = a.addForm.Update(msg)
		return a, cmd
	}

	return a, nil
}

func (a App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch a.mode {
	case modePushForm:
		return a.handlePushFormKey(msg)
	case modeAddForm:
		return a.handleAddFormKey(msg)
	}
	switch a.inputMode {
	case inputFilter:
		return a.handleFilterKey(msg)
	default:
		return a.handleBrowserKey(msg)
	}
}

func (a App) handleBrowserKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Quit):
		return a, tea.Quit

	case key.Matches(msg, a.keys.Down):
		if a.cursor < a.activeSection().NumRows()-1 {
			a.cursor++
		}

	case key.Matches(msg, a.keys.Up):
		if a.cursor > 0 {
			a.cursor--
		}

	case key.Matches(msg, a.keys.Top):
		a.cursor = 0

	case key.Matches(msg, a.keys.Bottom):
		if n := a.activeSection().NumRows(); n > 0 {
			a.cursor = n - 1
		}

	case key.Matches(msg, a.keys.NextTab):
		a.activeTab = (a.activeTab + 1) % len(a.sections)
		a.cursor = 0
		a.refreshSections()

	case key.Matches(msg, a.keys.PrevTab):
		a.activeTab = (a.activeTab - 1 + len(a.sections)) % len(a.sections)
		a.cursor = 0
		a.refreshSections()

	case key.Matches(msg, a.keys.Toggle):
		tasks := a.activeSection().Tasks()
		if a.cursor < len(tasks) {
			t := &tasks[a.cursor]
			return a, ToggleTaskCmd(t)
		}

	case key.Matches(msg, a.keys.Filter):
		a.inputMode = inputFilter
		a.input = ""
		return a, nil

	case key.Matches(msg, a.keys.AddTask):
		targets, defaultIdx := buildTargets(a.ctx.Config)
		showPush := a.ctx.Config.CalDAV.URL != ""
		defaultPush := a.ctx.Config.CalDAV.AutoPush
		a.addForm = addform.New(targets, defaultIdx, defaultPush, showPush)
		a.mode = modeAddForm
		return a, nil

	case key.Matches(msg, a.keys.Push):
		tasks := a.activeSection().Tasks()
		if a.cursor < len(tasks) && a.ctx.Config.CalDAV.URL != "" {
			t := &tasks[a.cursor]
			a.pushForm = pushform.New(t)
			a.mode = modePushForm
		} else if a.ctx.Config.CalDAV.URL == "" {
			a.message = "CalDAV not configured"
		}

	case key.Matches(msg, a.keys.ToggleView):
		if ts, ok := a.activeSection().(*tasksection.Model); ok {
			ts.ToggleGrouped()
		}

	case key.Matches(msg, a.keys.Reload):
		a.loading = true
		return a, LoadTasksCmd(a.ctx.VaultPath())
	}

	return a, nil
}

func (a App) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Enter):
		a.inputMode = inputNone
	case key.Matches(msg, a.keys.Escape):
		a.input = ""
		a.inputMode = inputNone
		a.applySearch()
	case key.Matches(msg, a.keys.Backspace):
		if len(a.input) > 0 {
			a.input = a.input[:len(a.input)-1]
		}
		a.applySearch()
	default:
		if len(msg.Runes) > 0 {
			a.input += string(msg.Runes)
			a.applySearch()
		}
	}
	return a, nil
}

func (a App) handleAddFormKey(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	a.addForm, cmd = a.addForm.Update(msg)

	if a.addForm.Cancelled() {
		a.mode = modeBrowser
		a.message = ""
		return a, nil
	}

	if a.addForm.Submitted() {
		cfg := a.ctx.Config.Vault
		target := a.addForm.GetTarget()
		filePath := vault.ResolveTaskFile(cfg.Path, cfg.DailyNotesFolder, cfg.DailyNotesFormat, cfg.DefaultTaskFile, target)
		summary := a.addForm.GetSummary()
		due := a.addForm.GetDue()
		priority := a.addForm.GetPriority()
		status := a.addForm.GetStatus()
		push := a.addForm.GetPush()
		a.mode = modeBrowser
		return a, AddTaskWithMetaCmd(filePath, summary, due, priority, status, push, a.ctx.Config.CalDAV)
	}

	return a, cmd
}

// buildTargets returns the ordered list of routing targets and the index of the default.
func buildTargets(cfg config.Config) ([]string, int) {
	targets := append([]string{"daily", "default"}, cfg.Vault.ExtraTargets...)
	defaultIdx := 0
	for i, t := range targets {
		if t == cfg.Vault.AddTaskTarget {
			defaultIdx = i
			break
		}
	}
	return targets, defaultIdx
}

func (a App) handlePushFormKey(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	a.pushForm, cmd = a.pushForm.Update(msg)

	if a.pushForm.Cancelled() {
		a.mode = modeBrowser
		a.message = ""
		return a, nil
	}

	if a.pushForm.Submitted() {
		tasks := a.activeSection().Tasks()
		if a.cursor < len(tasks) {
			t := &tasks[a.cursor]
			t.Description = a.pushForm.GetSummary()
			due := a.pushForm.GetDue()
			priority := a.pushForm.GetPriority()
			status := a.pushForm.GetStatus()

			uid, err := caldav.PushTask(a.ctx.Config.CalDAV, t, due, priority, status)
			if err != nil {
				a.message = "CalDAV push error: " + err.Error()
			} else {
				t.CalDAVUID = uid
				a.syncBack(t)
				// Persist UID to frontmatter for single-task files
				_ = vault.WriteFrontmatterUID(t.Source.FilePath, uid)
				a.message = "Pushed to CalDAV: " + t.Description
			}
		}
		a.mode = modeBrowser
		return a, nil
	}

	return a, cmd
}

func (a *App) activeSection() section.Section {
	return a.sections[a.activeTab]
}

func (a *App) refreshSections() {
	for _, s := range a.sections {
		s.SetTasks(a.allTasks)
	}
	if a.cursor >= a.activeSection().NumRows() && a.activeSection().NumRows() > 0 {
		a.cursor = a.activeSection().NumRows() - 1
	}
}

func (a *App) applySearch() {
	// Reset and re-filter with search term
	for _, s := range a.sections {
		if ts, ok := s.(*tasksection.Model); ok {
			ts.SetSearch(a.input)
			ts.SetTasks(a.allTasks)
		}
	}
	if a.cursor >= a.activeSection().NumRows() && a.activeSection().NumRows() > 0 {
		a.cursor = a.activeSection().NumRows() - 1
	}
}

func (a *App) syncBack(t *task.Task) {
	for i := range a.allTasks {
		if a.allTasks[i].Source == t.Source {
			a.allTasks[i].Status = t.Status
			a.allTasks[i].CalDAVUID = t.CalDAVUID
			a.allTasks[i].Description = t.Description
			break
		}
	}
}

func (a App) View() string {
	if a.ctx.VaultPath() == "" {
		return "No vault path configured. Set it in ~/.config/obia/config.toml\n\nPress q to quit."
	}

	w := a.ctx.Width
	if w < 1 {
		w = 80
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("  Obia"))
	b.WriteString("\n")

	// Tab bar
	b.WriteString(a.renderTabBar(w))
	b.WriteString("\n")

	// Task list
	listHeight := a.ctx.Height - 7
	if listHeight < 1 {
		listHeight = 10
	}

	if a.loading {
		b.WriteString("  Loading tasks...\n")
	} else {
		b.WriteString(a.activeSection().View(w, listHeight, a.cursor, true))
	}

	// Input line / overlay forms
	if a.inputMode == inputFilter {
		b.WriteString(filterPromptStyle.Render("/") + a.input + "█\n")
	} else if a.mode == modePushForm {
		b.WriteString("\n")
		b.WriteString(a.pushForm.View())
		b.WriteString("\n")
	} else if a.mode == modeAddForm {
		b.WriteString("\n")
		b.WriteString(a.addForm.View())
		b.WriteString("\n")
	}

	// Message
	if a.message != "" {
		b.WriteString(messageStyle.Render("  "+a.message) + "\n")
	}

	// Status bar
	b.WriteString(statusBarStyle.Width(w).Render(keys.BrowserHelp()))

	return b.String()
}

func (a App) renderTabBar(width int) string {
	var tabs []string
	for i, s := range a.sections {
		name := s.Title()
		count := s.NumRows()
		label := fmt.Sprintf("%s(%d)", name, count)
		if i == a.activeTab {
			tabs = append(tabs, activeTabStyle.Render("["+label+"]"))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(" "+label+" "))
		}
	}
	bar := strings.Join(tabs, "")
	return tabBarStyle.Width(width).Render(bar)
}
