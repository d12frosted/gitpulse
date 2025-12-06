package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/d12frosted/gitpulse/internal/config"
	"github.com/d12frosted/gitpulse/internal/git"
)

// Messages
type statusUpdatedMsg struct {
	index  int
	status *git.RepoStatus
}

type fetchCompleteMsg struct {
	index int
	err   error
}

type pullCompleteMsg struct {
	index int
	err   error
}

type pushCompleteMsg struct {
	index int
	err   error
}

type fetchAllCompleteMsg struct{}

// Model
type Model struct {
	repos       []config.RepoConfig
	statuses    []*git.RepoStatus
	cursor      int
	spinner     spinner.Model
	width       int
	height      int
	fetchingAll bool
	grouped     bool
	quitting    bool
}

func NewModel(repos []config.RepoConfig) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	statuses := make([]*git.RepoStatus, len(repos))
	for i, repo := range repos {
		statuses[i] = &git.RepoStatus{
			Path: repo.Path,
			Name: repo.Name,
		}
	}

	return Model{
		repos:    repos,
		statuses: statuses,
		spinner:  s,
		grouped:  true,
	}
}

// statusPriority returns a sort priority for a repo status
// Lower values appear first when grouped
func statusPriority(s *git.RepoStatus) int {
	if s.Error != nil {
		return 0 // Errors first
	}
	if s.NeedsPull() {
		return 1 // Needs pull (behind)
	}
	if s.NeedsPush() {
		return 2 // Needs push (ahead)
	}
	if s.IsSynced() {
		return 3 // Synced
	}
	return 4 // No upstream
}

// displayOrder returns indices in display order (sorted if grouped)
func (m *Model) displayOrder() []int {
	indices := make([]int, len(m.statuses))
	for i := range indices {
		indices[i] = i
	}

	if m.grouped {
		sort.Slice(indices, func(a, b int) bool {
			pa := statusPriority(m.statuses[indices[a]])
			pb := statusPriority(m.statuses[indices[b]])
			if pa != pb {
				return pa < pb
			}
			// Same priority: sort by name
			return m.statuses[indices[a]].Name < m.statuses[indices[b]].Name
		})
	}

	return indices
}

// selectedIndex returns the actual repo index for the current cursor position
func (m *Model) selectedIndex() int {
	return m.displayOrder()[m.cursor]
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.spinner.Tick}

	// Refresh all statuses on start
	for i, repo := range m.repos {
		cmds = append(cmds, m.refreshStatus(i, repo))
	}

	return tea.Batch(cmds...)
}

func (m *Model) refreshStatus(index int, repo config.RepoConfig) tea.Cmd {
	return func() tea.Msg {
		status := git.GetStatus(repo.Path, repo.Name)
		return statusUpdatedMsg{index: index, status: status}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.repos)-1 {
				m.cursor++
			}

		case "f":
			// Fetch all
			if !m.fetchingAll {
				m.fetchingAll = true
				cmds := make([]tea.Cmd, 0, len(m.repos))
				for i := range m.repos {
					m.statuses[i].Fetching = true
					cmds = append(cmds, m.fetchRepo(i))
				}
				return m, tea.Batch(cmds...)
			}

		case "enter", " ":
			// Fetch + pull current repo
			idx := m.selectedIndex()
			if !m.statuses[idx].Fetching && !m.statuses[idx].Rebasing {
				m.statuses[idx].Fetching = true
				m.statuses[idx].LastMessage = ""
				return m, m.fetchAndPull(idx)
			}

		case "p":
			// Push current repo
			idx := m.selectedIndex()
			if !m.statuses[idx].Pushing && m.statuses[idx].NeedsPush() {
				m.statuses[idx].Pushing = true
				m.statuses[idx].LastMessage = ""
				return m, m.pushRepo(idx)
			}

		case "r":
			// Refresh all statuses
			cmds := make([]tea.Cmd, 0, len(m.repos))
			for i, repo := range m.repos {
				cmds = append(cmds, m.refreshStatus(i, repo))
			}
			return m, tea.Batch(cmds...)

		case "g":
			// Toggle grouping by status
			m.grouped = !m.grouped
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case statusUpdatedMsg:
		if msg.index < len(m.statuses) {
			// Preserve operation states
			fetching := m.statuses[msg.index].Fetching
			rebasing := m.statuses[msg.index].Rebasing
			pushing := m.statuses[msg.index].Pushing
			lastMsg := m.statuses[msg.index].LastMessage

			m.statuses[msg.index] = msg.status
			m.statuses[msg.index].Fetching = fetching
			m.statuses[msg.index].Rebasing = rebasing
			m.statuses[msg.index].Pushing = pushing
			m.statuses[msg.index].LastMessage = lastMsg
		}

	case fetchCompleteMsg:
		if msg.index < len(m.statuses) {
			m.statuses[msg.index].Fetching = false
			if msg.err != nil {
				m.statuses[msg.index].LastMessage = fmt.Sprintf("fetch failed: %v", msg.err)
			}
		}
		// Check if all fetches are done
		allDone := true
		for _, s := range m.statuses {
			if s.Fetching {
				allDone = false
				break
			}
		}
		if allDone {
			m.fetchingAll = false
		}
		// Refresh status after fetch
		return m, m.refreshStatus(msg.index, m.repos[msg.index])

	case pullCompleteMsg:
		if msg.index < len(m.statuses) {
			m.statuses[msg.index].Rebasing = false
			if msg.err != nil {
				m.statuses[msg.index].LastMessage = fmt.Sprintf("pull failed: %v", msg.err)
			} else {
				m.statuses[msg.index].LastMessage = "synced"
			}
		}
		return m, m.refreshStatus(msg.index, m.repos[msg.index])

	case pushCompleteMsg:
		if msg.index < len(m.statuses) {
			m.statuses[msg.index].Pushing = false
			if msg.err != nil {
				m.statuses[msg.index].LastMessage = fmt.Sprintf("push failed: %v", msg.err)
			} else {
				m.statuses[msg.index].LastMessage = "pushed"
			}
		}
		return m, m.refreshStatus(msg.index, m.repos[msg.index])
	}

	return m, nil
}

func (m *Model) fetchRepo(index int) tea.Cmd {
	path := m.repos[index].Path
	return func() tea.Msg {
		err := git.Fetch(path)
		return fetchCompleteMsg{index: index, err: err}
	}
}

func (m *Model) fetchAndPull(index int) tea.Cmd {
	path := m.repos[index].Path
	return func() tea.Msg {
		// First fetch
		if err := git.Fetch(path); err != nil {
			return pullCompleteMsg{index: index, err: err}
		}
		// Then pull with rebase
		err := git.Pull(path)
		return pullCompleteMsg{index: index, err: err}
	}
}

func (m *Model) pushRepo(index int) tea.Cmd {
	path := m.repos[index].Path
	return func() tea.Msg {
		err := git.Push(path)
		return pushCompleteMsg{index: index, err: err}
	}
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	selectedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212"))

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	syncedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("42"))

	aheadStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214"))

	behindStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("204"))

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Italic(true)

	// Title
	b.WriteString(titleStyle.Render("  gitpulse"))
	b.WriteString("\n\n")

	// Calculate max name length for alignment
	maxNameLen := 0
	for _, repo := range m.repos {
		if len(repo.Name) > maxNameLen {
			maxNameLen = len(repo.Name)
		}
	}

	// Repos list
	order := m.displayOrder()
	for displayIdx, repoIdx := range order {
		status := m.statuses[repoIdx]
		cursor := "  "
		style := normalStyle
		if displayIdx == m.cursor {
			cursor = "▸ "
			style = selectedStyle
		}

		// Name (padded)
		name := fmt.Sprintf("%-*s", maxNameLen, status.Name)

		// Status indicator
		var statusStr string
		if status.Error != nil {
			statusStr = errorStyle.Render(" ✗ " + status.Error.Error())
		} else if status.Fetching {
			statusStr = dimStyle.Render(" " + m.spinner.View() + " fetching...")
		} else if status.Rebasing {
			statusStr = dimStyle.Render(" " + m.spinner.View() + " rebasing...")
		} else if status.Pushing {
			statusStr = dimStyle.Render(" " + m.spinner.View() + " pushing...")
		} else if !status.HasUpstream {
			statusStr = dimStyle.Render(" ○ no upstream")
		} else if status.IsSynced() {
			statusStr = syncedStyle.Render(" ✓ synced")
		} else {
			var parts []string
			if status.Ahead > 0 {
				parts = append(parts, aheadStyle.Render(fmt.Sprintf("↑%d", status.Ahead)))
			}
			if status.Behind > 0 {
				parts = append(parts, behindStyle.Render(fmt.Sprintf("↓%d", status.Behind)))
			}
			statusStr = " " + strings.Join(parts, " ")
		}

		// Branch + dirty indicator
		branch := ""
		if status.Branch != "" {
			dirty := ""
			if status.Dirty {
				dirty = aheadStyle.Render("*")
			}
			branch = dimStyle.Render(" ["+status.Branch+"]") + dirty
		}

		// Last message (operation feedback)
		msg := ""
		if status.LastMessage != "" {
			msg = " " + messageStyle.Render("← "+status.LastMessage)
		}

		// Last commit info
		commitInfo := ""
		if status.CommitSubject != "" && status.LastMessage == "" {
			subject := status.CommitSubject
			if len(subject) > 30 {
				subject = subject[:27] + "..."
			}
			commitInfo = " " + dimStyle.Render(status.CommitAge+": "+subject)
		}

		line := cursor + style.Render(name) + branch + statusStr + msg + commitInfo
		b.WriteString(line + "\n")
	}

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := "\n" + helpStyle.Render("  f: fetch all • enter: fetch+rebase • p: push • r: refresh • g: group • q: quit")
	b.WriteString(help)

	return b.String()
}
