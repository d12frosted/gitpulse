package ui

import (
	"fmt"
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
	}
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
			if !m.statuses[m.cursor].Fetching && !m.statuses[m.cursor].Rebasing {
				m.statuses[m.cursor].Fetching = true
				m.statuses[m.cursor].LastMessage = ""
				return m, m.fetchAndPull(m.cursor)
			}

		case "p":
			// Push current repo
			if !m.statuses[m.cursor].Pushing && m.statuses[m.cursor].NeedsPush() {
				m.statuses[m.cursor].Pushing = true
				m.statuses[m.cursor].LastMessage = ""
				return m, m.pushRepo(m.cursor)
			}

		case "r":
			// Refresh all statuses
			cmds := make([]tea.Cmd, 0, len(m.repos))
			for i, repo := range m.repos {
				cmds = append(cmds, m.refreshStatus(i, repo))
			}
			return m, tea.Batch(cmds...)
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
	for i, status := range m.statuses {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
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

		// Branch
		branch := ""
		if status.Branch != "" {
			branch = dimStyle.Render(" [" + status.Branch + "]")
		}

		// Last message
		msg := ""
		if status.LastMessage != "" {
			msg = " " + messageStyle.Render("← "+status.LastMessage)
		}

		line := cursor + style.Render(name) + branch + statusStr + msg
		b.WriteString(line + "\n")
	}

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := "\n" + helpStyle.Render("  f: fetch all • enter: fetch+rebase • p: push • r: refresh • q: quit")
	b.WriteString(help)

	return b.String()
}
