package hermessection

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hawkaii/obia/internal/hermes"
	"github.com/hawkaii/obia/internal/tui/components/section"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true).
			Padding(0, 1)

	sectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A78BFA")).
			Bold(true)

	taskStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E2E8F0"))

	dsaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#34D399"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#64748B"))

	commitStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FBBF24"))
)

// ContextBriefingMsg is sent when the Hermes context is loaded.
type ContextBriefingMsg struct {
	Briefing *hermes.ContextBriefing
	Err      error
}

// Model implements section.Section as a context briefing pane.
type Model struct {
	vaultPath string
	repo      string
	briefing  *hermes.ContextBriefing
	err       error
	loading   bool
}

var _ section.Section = (*Model)(nil)

func New(vaultPath, repo string) *Model {
	return &Model{
		vaultPath: vaultPath,
		repo:      repo,
		briefing:  nil,
		loading:   true,
	}
}

func (m *Model) Title() string              { return "Context" }
func (m *Model) Tasks() []hermes.DSATask    { return nil }
func (m *Model) NumRows() int               { return 1 }
func (m *Model) SetSearch(query string)     {}
func (m *Model) SetTasks(_ []interface{})   {}
func (m *Model) SetGrouped(v bool)          {}
func (m *Model) IsGrouped() bool            { return false }
func (m *Model) ToggleGrouped()             {}
func (m *Model) SetWarning(w string)        {}

func (m *Model) SetBriefing(b *hermes.ContextBriefing) {
	m.briefing = b
	m.loading = false
	m.err = nil
}

func (m *Model) SetError(err error) {
	m.err = err
	m.loading = false
}

func (m *Model) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	switch msg := msg.(type) {
	case ContextBriefingMsg:
		if msg.Err != nil {
			m.SetError(msg.Err)
		} else {
			m.SetBriefing(msg.Briefing)
		}
	}
	return m, nil
}

func (m *Model) Loading() bool { return m.loading }

func (m *Model) View(width, height, cursor int, selected bool) string {
	if m.loading {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			dimStyle.Render("⏳ Loading session context…"))
	}
	if m.err != nil {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			dimStyle.Render(fmt.Sprintf("⚠️ %v", m.err)))
	}
	if m.briefing == nil {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			dimStyle.Render("No context available. Start coding and I'll track it."))
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("⚡ Session Context") + "\n\n")

	// Repo + DSA streak header
	repo := m.briefing.Repo
	if repo == "" {
		repo = "unknown"
	}
	streak := ""
	if m.briefing.DSA.Solved > 0 {
		streak = fmt.Sprintf("  DSA: %d/%d 🔥", m.briefing.DSA.Solved, m.briefing.DSA.Total)
	}
	b.WriteString(sectionStyle.Render(fmt.Sprintf("📁 %s%s", repo, streak)))
	b.WriteString("\n\n")

	// Uncommitted changes
	if len(m.briefing.Unstaged) > 0 {
		b.WriteString(dimStyle.Render("📦 Uncommitted:"))
		b.WriteString("\n")
		for _, f := range m.briefing.Unstaged[:min(len(m.briefing.Unstaged), 5)] {
			mark := " "
			if f.Status == "M" || f.Status == "?" {
				mark = "◆"
			}
			b.WriteString(fmt.Sprintf("   %s %s\n", dimStyle.Render(mark), taskStyle.Render(f.File)))
		}
		b.WriteString("\n")
	}

	// Vault logs
	if len(m.briefing.VaultLogs) > 0 {
		b.WriteString(sectionStyle.Render("📋 Last session:"))
		b.WriteString("\n")
		log := m.briefing.VaultLogs[0]
		b.WriteString(fmt.Sprintf("   %s\n", dimStyle.Render(log.File)))

		// Extract title from frontmatter preview
		preview := log.Preview
		if idx := strings.Index(preview, "title:"); idx >= 0 {
			titleLine := preview[idx:]
			if end := strings.Index(titleLine, "\n"); end >= 0 {
				titleLine = titleLine[:end]
				titleLine = strings.TrimPrefix(titleLine, "title:")
				titleLine = strings.Trim(titleLine, ` "'`)
				b.WriteString(fmt.Sprintf("   %s\n", taskStyle.Render(titleLine)))
			}
		}
		b.WriteString("\n")
	}

	// Recent commits
	if len(m.briefing.Recent) > 0 {
		for _, rc := range m.briefing.Recent {
			if len(rc.Commits) > 0 {
				b.WriteString(sectionStyle.Render(fmt.Sprintf("🔨 Recent %s:", rc.Repo)))
				b.WriteString("\n")
				for _, c := range rc.Commits[:min(len(rc.Commits), 3)] {
					b.WriteString(fmt.Sprintf("   %s\n", commitStyle.Render(c)))
				}
				b.WriteString("\n")
			}
		}
	}

	// DSA
	if m.briefing.DSA.Total > 0 {
		b.WriteString(sectionStyle.Render("🏋️ DSA Progress:"))
		b.WriteString("\n")
		for _, p := range m.briefing.DSA.Problems[:min(len(m.briefing.DSA.Problems), 5)] {
			status := " "
			if p.Solved {
				status = "✅"
			}
			b.WriteString(fmt.Sprintf("   %s %s\n", status, taskStyle.Render(getName(p.Name))))
		}
		if m.briefing.DSA.Solved > 0 {
			b.WriteString(fmt.Sprintf("   %s\n", dimStyle.Render(fmt.Sprintf("Total: %d/%d solved", m.briefing.DSA.Solved, m.briefing.DSA.Total))))
		}
		b.WriteString("\n")
	}

	// Status line
	b.WriteString(dimStyle.Render(fmt.Sprintf("Updated: %s · %dms", m.briefing.Timestamp[:19], m.briefing.Meta.Ms)))

	return b.String()
}

func getName(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return strings.ReplaceAll(parts[len(parts)-1], "-", " ")
	}
	return url
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
