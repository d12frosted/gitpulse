package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name        string
	Border      lipgloss.Color
	Title       lipgloss.Color
	RepoName    lipgloss.Color
	Selected    lipgloss.Color
	Branch      lipgloss.Color
	Synced      lipgloss.Color
	Ahead       lipgloss.Color
	Behind      lipgloss.Color
	Error       lipgloss.Color
	Dim         lipgloss.Color
	HelpKey     lipgloss.Color
	HelpText    lipgloss.Color
	NoRemote    lipgloss.Color
	Spinner     lipgloss.Color
}

var Themes = map[string]Theme{
	"dracula": {
		Name:     "dracula",
		Border:   lipgloss.Color("#6272a4"),
		Title:    lipgloss.Color("#ff79c6"),
		RepoName: lipgloss.Color("#f8f8f2"),
		Selected: lipgloss.Color("#ff79c6"),
		Branch:   lipgloss.Color("#6272a4"),
		Synced:   lipgloss.Color("#50fa7b"),
		Ahead:    lipgloss.Color("#f1fa8c"),
		Behind:   lipgloss.Color("#ff5555"),
		Error:    lipgloss.Color("#ff5555"),
		Dim:      lipgloss.Color("#44475a"),
		HelpKey:  lipgloss.Color("#bd93f9"),
		HelpText: lipgloss.Color("#6272a4"),
		NoRemote: lipgloss.Color("#6272a4"),
		Spinner:  lipgloss.Color("#ff79c6"),
	},
	"nord": {
		Name:     "nord",
		Border:   lipgloss.Color("#4c566a"),
		Title:    lipgloss.Color("#88c0d0"),
		RepoName: lipgloss.Color("#eceff4"),
		Selected: lipgloss.Color("#88c0d0"),
		Branch:   lipgloss.Color("#4c566a"),
		Synced:   lipgloss.Color("#a3be8c"),
		Ahead:    lipgloss.Color("#ebcb8b"),
		Behind:   lipgloss.Color("#bf616a"),
		Error:    lipgloss.Color("#bf616a"),
		Dim:      lipgloss.Color("#3b4252"),
		HelpKey:  lipgloss.Color("#b48ead"),
		HelpText: lipgloss.Color("#4c566a"),
		NoRemote: lipgloss.Color("#4c566a"),
		Spinner:  lipgloss.Color("#88c0d0"),
	},
	"catppuccin": {
		Name:     "catppuccin",
		Border:   lipgloss.Color("#585b70"),
		Title:    lipgloss.Color("#cba6f7"),
		RepoName: lipgloss.Color("#cdd6f4"),
		Selected: lipgloss.Color("#cba6f7"),
		Branch:   lipgloss.Color("#6c7086"),
		Synced:   lipgloss.Color("#a6e3a1"),
		Ahead:    lipgloss.Color("#f9e2af"),
		Behind:   lipgloss.Color("#f38ba8"),
		Error:    lipgloss.Color("#f38ba8"),
		Dim:      lipgloss.Color("#45475a"),
		HelpKey:  lipgloss.Color("#89b4fa"),
		HelpText: lipgloss.Color("#6c7086"),
		NoRemote: lipgloss.Color("#6c7086"),
		Spinner:  lipgloss.Color("#cba6f7"),
	},
	"gruvbox": {
		Name:     "gruvbox",
		Border:   lipgloss.Color("#665c54"),
		Title:    lipgloss.Color("#fe8019"),
		RepoName: lipgloss.Color("#ebdbb2"),
		Selected: lipgloss.Color("#fe8019"),
		Branch:   lipgloss.Color("#7c6f64"),
		Synced:   lipgloss.Color("#b8bb26"),
		Ahead:    lipgloss.Color("#fabd2f"),
		Behind:   lipgloss.Color("#fb4934"),
		Error:    lipgloss.Color("#fb4934"),
		Dim:      lipgloss.Color("#504945"),
		HelpKey:  lipgloss.Color("#d3869b"),
		HelpText: lipgloss.Color("#7c6f64"),
		NoRemote: lipgloss.Color("#7c6f64"),
		Spinner:  lipgloss.Color("#fe8019"),
	},
	"tokyonight": {
		Name:     "tokyonight",
		Border:   lipgloss.Color("#3b4261"),
		Title:    lipgloss.Color("#7aa2f7"),
		RepoName: lipgloss.Color("#c0caf5"),
		Selected: lipgloss.Color("#7aa2f7"),
		Branch:   lipgloss.Color("#565f89"),
		Synced:   lipgloss.Color("#9ece6a"),
		Ahead:    lipgloss.Color("#e0af68"),
		Behind:   lipgloss.Color("#f7768e"),
		Error:    lipgloss.Color("#f7768e"),
		Dim:      lipgloss.Color("#292e42"),
		HelpKey:  lipgloss.Color("#bb9af7"),
		HelpText: lipgloss.Color("#565f89"),
		NoRemote: lipgloss.Color("#565f89"),
		Spinner:  lipgloss.Color("#7aa2f7"),
	},
	"mono": {
		Name:     "mono",
		Border:   lipgloss.Color("#666666"),
		Title:    lipgloss.Color("#ffffff"),
		RepoName: lipgloss.Color("#ffffff"),
		Selected: lipgloss.Color("#ffffff"),
		Branch:   lipgloss.Color("#888888"),
		Synced:   lipgloss.Color("#aaaaaa"),
		Ahead:    lipgloss.Color("#ffffff"),
		Behind:   lipgloss.Color("#ffffff"),
		Error:    lipgloss.Color("#ff0000"),
		Dim:      lipgloss.Color("#444444"),
		HelpKey:  lipgloss.Color("#ffffff"),
		HelpText: lipgloss.Color("#666666"),
		NoRemote: lipgloss.Color("#666666"),
		Spinner:  lipgloss.Color("#ffffff"),
	},
	"jrpg-dark": {
		Name:     "jrpg-dark",
		Border:   lipgloss.Color("#7ec8e3"), // Cyan - MP/magic
		Title:    lipgloss.Color("#ffd866"), // Gold - highlighted
		RepoName: lipgloss.Color("#f4e4bc"), // Parchment - old RPG text
		Selected: lipgloss.Color("#ffd866"), // Gold
		Branch:   lipgloss.Color("#7ec8e3"), // Cyan
		Synced:   lipgloss.Color("#98d982"), // Cure Green - healing
		Ahead:    lipgloss.Color("#ffb347"), // Amber - warning
		Behind:   lipgloss.Color("#ff6b6b"), // Damage Red
		Error:    lipgloss.Color("#ff6b6b"), // Damage Red
		Dim:      lipgloss.Color("#8b9bb4"), // Slate - muted
		HelpKey:  lipgloss.Color("#c9b8ff"), // Lavender - magic glow
		HelpText: lipgloss.Color("#8b9bb4"), // Slate
		NoRemote: lipgloss.Color("#7ec8e3"), // Cyan
		Spinner:  lipgloss.Color("#ffd866"), // Gold
	},
	"jrpg-light": {
		Name:     "jrpg-light",
		Border:   lipgloss.Color("#3d7a99"), // Ocean Teal
		Title:    lipgloss.Color("#c4880d"), // Dark Gold
		RepoName: lipgloss.Color("#2d3a4f"), // Deep Navy
		Selected: lipgloss.Color("#c4880d"), // Dark Gold
		Branch:   lipgloss.Color("#3d7a99"), // Ocean Teal
		Synced:   lipgloss.Color("#2d8a4e"), // Forest Green
		Ahead:    lipgloss.Color("#d4780a"), // Burnt Orange
		Behind:   lipgloss.Color("#c13b3b"), // Crimson
		Error:    lipgloss.Color("#c13b3b"), // Crimson
		Dim:      lipgloss.Color("#8b96a3"), // Storm Grey
		HelpKey:  lipgloss.Color("#7b5caa"), // Royal Purple
		HelpText: lipgloss.Color("#8b96a3"), // Storm Grey
		NoRemote: lipgloss.Color("#3d7a99"), // Ocean Teal
		Spinner:  lipgloss.Color("#c4880d"), // Dark Gold
	},
}

var DefaultTheme = "dracula"

func GetTheme(name string) Theme {
	if theme, ok := Themes[name]; ok {
		return theme
	}
	return Themes[DefaultTheme]
}

func ThemeNames() []string {
	names := make([]string, 0, len(Themes))
	for name := range Themes {
		names = append(names, name)
	}
	return names
}
