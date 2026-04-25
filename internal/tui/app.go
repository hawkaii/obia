package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hawkaii/obia/internal/config"
	"github.com/hawkaii/obia/internal/task"
	"github.com/hawkaii/obia/internal/tui/components/addform"
	"github.com/hawkaii/obia/internal/tui/components/editform"
	"github.com/hawkaii/obia/internal/tui/components/section"
	"github.com/hawkaii/obia/internal/tui/components/tasksection"
	appctx "github.com/hawkaii/obia/internal/tui/context"
	"github.com/hawkaii/obia/internal/tui/keys"
	"github.com/hawkaii/obia/internal/vault"
)

type appMode int

const (
	modeBrowser appMode = iota
	modeAddForm
	modeEditForm
)

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

	allTasks  []task.Task
	sections  []section.Section
	activeTab int
	cursor    int
	loading   bool

	spinner spinner.Model

	addForm      addform.Model
	addFormTask  *task.Task // non-nil when p opens addform on an existing task
	editForm     editform.Model
	editFormTask *task.Task
}

func NewApp(cfg config.Config) App {
	ctx := appctx.New(cfg)

	vp := cfg.Vault.Path
	dfmt := cfg.Vault.DailyNotesFormat
	dfs := cfg.Vault.Folders
	if len(dfs) == 0 {
		dfs = []string{cfg.Vault.DailyNotesFolder}
	}

	sections := []section.Section{
		tasksection.New("Tasks",   vp, dfs, dfmt, tasksection.FilterOpen),
		tasksection.New("Daily",   vp, dfs, dfmt, tasksection.FilterDailyFolders),
		tasksection.New("Weekly",  vp, dfs, dfmt, tasksection.FilterWeekly),
		tasksection.New("Overdue", vp, dfs, dfmt, tasksection.FilterOverdue),
		tasksection.New("CalDAV",  vp, dfs, dfmt, tasksection.FilterCalDAV),
	}

	if cfg.UI.Grouped {
		for _, s := range sections {
			if ts, ok := s.(*tasksection.Model); ok {
				ts.SetGrouped(true)
			}
		}
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))

	return App{
		ctx:      ctx,
		keys:     keys.DefaultKeyMap,
		mode:     modeBrowser,
		sections: sections,
		loading:  true,
		spinner:  sp,
	}
}

func (a App) Init() tea.Cmd {
	return tea.Batch(
		LoadTasksCmd(a.ctx.VaultPath(), a.ctx.Config.Vault.TaskFilesFolder),
		a.spinner.Tick,
	)
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
		} else if msg.CalDAVErr != nil {
			msg.Task.Toggle()
			a.syncBack(msg.Task)
			a.refreshSections()
			a.message = "Toggled · CalDAV push failed: " + msg.CalDAVErr.Error()
		} else {
			msg.Task.Toggle()
			a.syncBack(msg.Task)
			a.refreshSections()
			a.message = "Toggled"
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
		return a, LoadTasksCmd(a.ctx.VaultPath(), a.ctx.Config.Vault.TaskFilesFolder)

	case PullCalDAVMsg:
		if msg.Err != nil {
			a.message = "Pull failed: " + msg.Err.Error()
		} else {
			a.message = fmt.Sprintf("Pull complete — %d updated, %d new", msg.Updated, msg.Created)
			if msg.Notify != "" {
				a.message += " · " + msg.Notify
			}
			return a, LoadTasksCmd(a.ctx.VaultPath(), a.ctx.Config.Vault.TaskFilesFolder)
		}

	case TaskEditedMsg:
		if msg.Err != nil {
			a.message = "Edit error: " + msg.Err.Error()
		} else if msg.Reload {
			if msg.CalDAVErr != nil {
				a.message = "Saved · CalDAV push failed: " + msg.CalDAVErr.Error()
			} else {
				a.message = "Task updated"
			}
			return a, LoadTasksCmd(a.ctx.VaultPath(), a.ctx.Config.Vault.TaskFilesFolder)
		} else {
			msg.Task.Description = msg.NewSummary
			a.syncBack(msg.Task)
			a.refreshSections()
			if msg.CalDAVErr != nil {
				a.message = "Saved · CalDAV push failed: " + msg.CalDAVErr.Error()
			} else {
				a.message = "Task updated"
			}
		}

	case CalDAVPushedMsg:
		if msg.Err != nil {
			a.message = "CalDAV push error: " + msg.Err.Error()
		} else {
			msg.Task.CalDAVUID = msg.UID
			a.syncBack(msg.Task)
			_ = vault.WriteFrontmatterUID(msg.Task.Source.FilePath, msg.UID)
			a.refreshSections()
			a.message = "Pushed to CalDAV: " + msg.Task.Description
		}

	case spinner.TickMsg:
		if a.loading {
			var cmd tea.Cmd
			a.spinner, cmd = a.spinner.Update(msg)
			return a, cmd
		}

	case tea.KeyMsg:
		return a.handleKey(msg)
	}

	if a.mode == modeAddForm {
		var cmd tea.Cmd
		a.addForm, cmd = a.addForm.Update(msg)
		return a, cmd
	}
	if a.mode == modeEditForm {
		var cmd tea.Cmd
		a.editForm, cmd = a.editForm.Update(msg)
		return a, cmd
	}

	return a, nil
}

func (a App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch a.mode {
	case modeAddForm:
		return a.handleAddFormKey(msg)
	case modeEditForm:
		return a.handleEditFormKey(msg)
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
			return a, ToggleTaskCmd(t, a.ctx.Config.CalDAV)
		}

	case key.Matches(msg, a.keys.Filter):
		a.inputMode = inputFilter
		a.input = ""
		return a, nil

	case key.Matches(msg, a.keys.AddTask):
		targets, defaultIdx := buildTargets(a.ctx.Config)
		showPush := a.ctx.Config.CalDAV.URL != ""
		defaultPush := a.ctx.Config.CalDAV.AutoPush
		a.addForm = addform.New("", targets, defaultIdx, defaultPush, showPush)
		a.addFormTask = nil
		a.mode = modeAddForm
		return a, nil

	case key.Matches(msg, a.keys.Push):
		tasks := a.activeSection().Tasks()
		if a.cursor >= len(tasks) {
			break
		}
		t := &tasks[a.cursor]
		if t.LinkedTaskFile != "" {
			a.message = "Already a CalDAV task — use R to pull updates"
			break
		}
		targets, defaultIdx := buildTargets(a.ctx.Config)
		showPush := a.ctx.Config.CalDAV.URL != ""
		a.addForm = addform.New(t.Description, targets, defaultIdx, true, showPush)
		a.addFormTask = t
		a.mode = modeAddForm
		return a, nil

	case key.Matches(msg, a.keys.Pull):
		if a.ctx.Config.CalDAV.URL == "" {
			a.message = "CalDAV not configured"
			break
		}
		a.message = "Pulling from CalDAV..."
		return a, PullCalDAVCmd(a.ctx.Config)

	case key.Matches(msg, a.keys.ToggleView):
		if ts, ok := a.activeSection().(*tasksection.Model); ok {
			ts.ToggleGrouped()
			a.ctx.Config.UI.Grouped = ts.IsGrouped()
			_ = config.Save(a.ctx.Config)
		}

	case key.Matches(msg, a.keys.EditTask):
		tasks := a.activeSection().Tasks()
		if a.cursor >= len(tasks) {
			break
		}
		t := &tasks[a.cursor]
		showPush := a.ctx.Config.CalDAV.URL != "" && t.CalDAVUID == ""
		a.editForm = editform.New(t, showPush)
		a.editFormTask = t
		a.mode = modeEditForm
		return a, nil

	case key.Matches(msg, a.keys.Reload):
		a.loading = true
		return a, LoadTasksCmd(a.ctx.VaultPath(), a.ctx.Config.Vault.TaskFilesFolder)
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
		summary := a.addForm.GetSummary()
		target := a.addForm.GetTarget()
		due := a.addForm.GetDue()
		priority := a.addForm.GetPriority()
		status := a.addForm.GetStatus()
		description := a.addForm.GetDescription()
		push := a.addForm.GetPush()
		cfg := a.ctx.Config

		a.mode = modeBrowser

		if a.addFormTask != nil {
			// p key path: link an existing plain task
			return a, LinkExistingTaskCmd(a.addFormTask, summary, description, due, priority, status, push, cfg)
		}

		// a key path: create new task
		vcfg := cfg.Vault
		filePath := vault.ResolveTaskFile(vcfg.Path, vcfg.DailyNotesFolder, vcfg.DailyNotesFormat, vcfg.DefaultTaskFile, target)
		return a, AddTaskWithMetaCmd(filePath, summary, description, due, priority, status, push, cfg)
	}

	return a, cmd
}

func (a App) handleEditFormKey(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	a.editForm, cmd = a.editForm.Update(msg)

	if a.editForm.Cancelled() {
		a.mode = modeBrowser
		a.message = ""
		return a, nil
	}

	if a.editForm.Submitted() {
		summary := a.editForm.GetSummary()
		due := a.editForm.GetDue()
		priority := a.editForm.GetPriority()
		status := a.editForm.GetStatus()
		description := a.editForm.GetDescription()
		push := a.editForm.GetPush()
		cfg := a.ctx.Config

		a.mode = modeBrowser
		return a, EditTaskCmd(a.editFormTask, summary, description, due, priority, status, push, cfg)
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

func (a App) View() string {
	if a.ctx.VaultPath() == "" {
		return "No vault path configured. Set it in ~/.config/obia/config.toml\n\nPress q to quit."
	}

	w := a.ctx.Width
	if w < 1 {
		w = 80
	}

	var b strings.Builder

	subtitle := "manage tasks like a god"
	innerW := lipgloss.Width(logo)
	if lipgloss.Width(subtitle) > innerW {
		innerW = lipgloss.Width(subtitle)
	}

	logoInner := lipgloss.JoinVertical(lipgloss.Left,
		logoStyle.Width(innerW). /*.Background(lipgloss.Color("170")).*/ Render(logo),
		logoSubtitleStyle.Width(innerW). /*.Background(lipgloss.Color("170")).*/ Render(subtitle),
	)
	logoBlock := lipgloss.NewStyle().
		// Background(lipgloss.Color("170")).
		Padding(1, 6, 1, 2).
		Render(logoInner)
	headerHeight := lipgloss.Height(logoBlock)
	tabsBlock := lipgloss.NewStyle().
		Width(w - lipgloss.Width(logoBlock)).
		Height(headerHeight).
		AlignVertical(lipgloss.Center).
		Render(a.renderTabs())
	b.WriteString(tabBarStyle.Width(w).Height(headerHeight).Render(
		lipgloss.JoinHorizontal(lipgloss.Bottom, tabsBlock, logoBlock),
	))
	b.WriteString("\n")

	listHeight := a.ctx.Height - headerHeight - 5
	if listHeight < 1 {
		listHeight = 10
	}

	if a.loading {
		msg := fmt.Sprintf("%s loading vault…", a.spinner.View())
		b.WriteString(lipgloss.Place(w, listHeight, lipgloss.Center, lipgloss.Center, msg))
	} else {
		b.WriteString(a.activeSection().View(w, listHeight, a.cursor, true))
	}

	if a.inputMode == inputFilter {
		b.WriteString(filterPromptStyle.Render("/") + a.input + "█\n")
	} else if a.mode == modeAddForm {
		b.WriteString("\n")
		b.WriteString(a.addForm.View())
		b.WriteString("\n")
	} else if a.mode == modeEditForm {
		b.WriteString("\n")
		b.WriteString(a.editForm.View())
		b.WriteString("\n")
	}

	if a.message != "" {
		b.WriteString(messageStyle.Render("  "+a.message) + "\n")
	}

	b.WriteString(statusBarStyle.Width(w).Render(keys.BrowserHelp()))

	return b.String()
}

func (a App) renderTabs() string {
	var tabs []string
	for i, s := range a.sections {
		name := s.Title()
		count := s.NumRows()
		label := fmt.Sprintf("%s(%d)", name, count)
		if i == a.activeTab {
			tabs = append(tabs, activeTabStyle.Render(label))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(label))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}
