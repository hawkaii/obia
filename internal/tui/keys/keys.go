package keys

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up         key.Binding
	Down       key.Binding
	Top        key.Binding
	Bottom     key.Binding
	NextTab    key.Binding
	PrevTab    key.Binding
	Toggle     key.Binding
	Push       key.Binding
	Filter     key.Binding
	AddTask    key.Binding
	Reload     key.Binding
	Pull       key.Binding
	ToggleView key.Binding
	Quit       key.Binding
	Escape     key.Binding
	Backspace  key.Binding
	Enter      key.Binding
}

var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("↓/j", "down"),
	),
	Top: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "top"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("G"),
		key.WithHelp("G", "bottom"),
	),
	NextTab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab"),
	),
	PrevTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev tab"),
	),
	Toggle: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "toggle"),
	),
	Push: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "push caldav"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	AddTask: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add task"),
	),
	Reload: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reload"),
	),
	Pull: key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "pull caldav"),
	),
	ToggleView: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "view"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
	Backspace: key.NewBinding(
		key.WithKeys("backspace"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
	),
}

// BrowserHelp returns the help text for the browser mode status bar.
func BrowserHelp() string {
	return "  ↑/k ↓/j navigate  enter: toggle  a: add  p: push  R: pull  /: filter  v: view  tab: switch  r: reload  q: quit"
}
