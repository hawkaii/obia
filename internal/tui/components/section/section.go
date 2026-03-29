package section

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hawkaii/obia/internal/task"
)

// Section is the interface every browser tab must implement.
type Section interface {
	Update(msg tea.Msg) (Section, tea.Cmd)
	View(width, height, cursor int, selected bool) string
	SetTasks(all []task.Task)
	Tasks() []task.Task
	NumRows() int
	Title() string
}
