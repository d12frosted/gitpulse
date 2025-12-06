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
	theme       Theme
}

func NewModel(repos []config.RepoConfig, themeName string) Model {
	theme := GetTheme(themeName)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Spinner)

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
		theme:    theme,
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

	// Use terminal width, with some padding
	width := m.width
	if width < 60 {
		width = 80
	}
	innerWidth := width - 4 // account for border + padding

	// Theme colors
	t := m.theme

	// Calculate column widths
	maxNameLen := 0
	maxBranchLen := 0
	for _, s := range m.statuses {
		if len(s.Name) > maxNameLen {
			maxNameLen = len(s.Name)
		}
		if len(s.Branch) > maxBranchLen {
			maxBranchLen = len(s.Branch)
		}
	}
	if maxBranchLen > 14 {
		maxBranchLen = 14
	}

	// Build repo lines
	var lines []string
	order := m.displayOrder()
	for displayIdx, repoIdx := range order {
		status := m.statuses[repoIdx]
		isSelected := displayIdx == m.cursor

		var parts []string

		// Cursor
		if isSelected {
			parts = append(parts, lipgloss.NewStyle().Foreground(t.Selected).Render("▸"))
		} else {
			parts = append(parts, " ")
		}

		// Name
		name := fmt.Sprintf("%-*s", maxNameLen, status.Name)
		if isSelected {
			parts = append(parts, lipgloss.NewStyle().Bold(true).Foreground(t.Selected).Render(name))
		} else {
			parts = append(parts, lipgloss.NewStyle().Foreground(t.RepoName).Render(name))
		}

		// Branch
		branch := status.Branch
		if len(branch) > maxBranchLen {
			branch = branch[:maxBranchLen-1] + "…"
		}
		branchStr := fmt.Sprintf("%-*s", maxBranchLen, branch)
		parts = append(parts, lipgloss.NewStyle().Foreground(t.Branch).Render(branchStr))

		// Dirty
		if status.Dirty {
			parts = append(parts, lipgloss.NewStyle().Bold(true).Foreground(t.Ahead).Render("*"))
		} else {
			parts = append(parts, " ")
		}

		// Status
		statusWidth := 12
		var statusStr string
		if status.Error != nil {
			errMsg := status.Error.Error()
			if len(errMsg) > statusWidth-2 {
				errMsg = errMsg[:statusWidth-3] + "…"
			}
			statusStr = lipgloss.NewStyle().Foreground(t.Error).Render(fmt.Sprintf("✗ %-*s", statusWidth-2, errMsg))
		} else if status.Fetching {
			statusStr = lipgloss.NewStyle().Foreground(t.Spinner).Render(m.spinner.View()+" fetch…")
			statusStr = fmt.Sprintf("%-*s", statusWidth, statusStr)
		} else if status.Rebasing {
			statusStr = lipgloss.NewStyle().Foreground(t.Spinner).Render(m.spinner.View()+" rebase…")
			statusStr = fmt.Sprintf("%-*s", statusWidth, statusStr)
		} else if status.Pushing {
			statusStr = lipgloss.NewStyle().Foreground(t.Spinner).Render(m.spinner.View()+" push…")
			statusStr = fmt.Sprintf("%-*s", statusWidth, statusStr)
		} else if !status.HasUpstream {
			statusStr = lipgloss.NewStyle().Foreground(t.NoRemote).Render(fmt.Sprintf("%-*s", statusWidth, "○ no remote"))
		} else if status.IsSynced() {
			statusStr = lipgloss.NewStyle().Bold(true).Foreground(t.Synced).Render(fmt.Sprintf("%-*s", statusWidth, "✓ synced"))
		} else {
			var statusParts []string
			if status.Ahead > 0 {
				statusParts = append(statusParts, lipgloss.NewStyle().Bold(true).Foreground(t.Ahead).Render(fmt.Sprintf("↑%d", status.Ahead)))
			}
			if status.Behind > 0 {
				statusParts = append(statusParts, lipgloss.NewStyle().Bold(true).Foreground(t.Behind).Render(fmt.Sprintf("↓%d", status.Behind)))
			}
			statusStr = strings.Join(statusParts, " ")
			// Pad to fixed width
			visWidth := lipgloss.Width(statusStr)
			if visWidth < statusWidth {
				statusStr += strings.Repeat(" ", statusWidth-visWidth)
			}
		}
		parts = append(parts, statusStr)

		// Commit info - use remaining space
		usedWidth := 1 + 1 + maxNameLen + 1 + maxBranchLen + 1 + 1 + statusWidth + 2
		remainingWidth := innerWidth - usedWidth
		if remainingWidth > 10 && status.CommitSubject != "" && status.Error == nil {
			age := status.CommitAge
			// Shorten age
			ageParts := strings.Split(age, " ")
			if len(ageParts) >= 2 {
				age = ageParts[0] + string(ageParts[1][0])
			}
			ageWidth := 5
			subjectWidth := remainingWidth - ageWidth - 1
			if subjectWidth > 0 {
				subject := status.CommitSubject
				if len(subject) > subjectWidth {
					subject = subject[:subjectWidth-1] + "…"
				}
				commitInfo := fmt.Sprintf("%*s %s", ageWidth, age, subject)
				parts = append(parts, lipgloss.NewStyle().Foreground(t.Dim).Render(commitInfo))
			}
		}

		line := strings.Join(parts, " ")
		lines = append(lines, line)
	}

	// Build help line
	helpItems := []struct{ key, desc string }{
		{"f", "fetch"},
		{"⏎", "sync"},
		{"p", "push"},
		{"r", "refresh"},
		{"g", "group"},
		{"q", "quit"},
	}
	var helpParts []string
	for _, item := range helpItems {
		key := lipgloss.NewStyle().Bold(true).Foreground(t.HelpKey).Render(item.key)
		desc := lipgloss.NewStyle().Foreground(t.HelpText).Render(item.desc)
		helpParts = append(helpParts, key+" "+desc)
	}
	helpLine := strings.Join(helpParts, "  ")

	// Combine content
	content := strings.Join(lines, "\n")

	// Create box style
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Padding(1, 2).
		Width(width - 2)

	// Title style
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Title).
		MarginBottom(1)

	// Final layout
	var b strings.Builder
	b.WriteString("\n")

	innerContent := titleStyle.Render("gitpulse") + "\n\n" + content + "\n\n" + helpLine
	b.WriteString(boxStyle.Render(innerContent))
	b.WriteString("\n")

	return b.String()
}
