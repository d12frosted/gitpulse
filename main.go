package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/d12frosted/gitpulse/internal/config"
	"github.com/d12frosted/gitpulse/internal/ui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		var notFound *config.ConfigNotFoundError
		if errors.As(err, &notFound) {
			handleMissingConfig()
			return
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(cfg.Repos) == 0 {
		fmt.Println("No repositories configured.")
		fmt.Printf("Add repositories to %s\n", config.ConfigPath())
		os.Exit(1)
	}

	repos := cfg.RepoConfigs()

	p := tea.NewProgram(
		ui.NewModel(repos),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleMissingConfig() {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	pathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("cyan"))

	fmt.Println()
	fmt.Println(titleStyle.Render("  gitpulse"))
	fmt.Println()
	fmt.Println("  Config file not found.")
	fmt.Println()
	fmt.Printf("  Expected location: %s\n", pathStyle.Render(config.ConfigPath()))
	fmt.Println()
	fmt.Println(dimStyle.Render("  Example config:"))
	fmt.Println()

	for _, line := range strings.Split(config.ExampleConfig(), "\n") {
		fmt.Printf("  %s\n", dimStyle.Render(line))
	}

	fmt.Println()
	fmt.Print("  Would you like to create a config now? [Y/n] ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "" && input != "y" && input != "yes" {
		fmt.Println()
		fmt.Println("  No config created. Exiting.")
		return
	}

	// Interactive config creation
	fmt.Println()
	fmt.Println("  Enter repository paths (one per line, empty line to finish):")
	fmt.Println()

	var repos []string
	for {
		fmt.Print("  > ")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)

		if line == "" {
			break
		}

		// Expand and validate path
		expanded := expandPath(line)
		if _, err := os.Stat(expanded); os.IsNotExist(err) {
			fmt.Printf("    %s does not exist, adding anyway\n", dimStyle.Render(line))
		}

		// Check if it's a git repo
		gitDir := filepath.Join(expanded, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			fmt.Printf("    %s is not a git repository, adding anyway\n", dimStyle.Render(line))
		}

		repos = append(repos, line)
	}

	if len(repos) == 0 {
		fmt.Println()
		fmt.Println("  No repositories added. Exiting.")
		return
	}

	cfg := &config.Config{Repos: repos}
	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "  Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("  Config saved to %s\n", pathStyle.Render(config.ConfigPath()))
	fmt.Println()
	fmt.Println("  Run gitpulse again to start monitoring your repos.")
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
