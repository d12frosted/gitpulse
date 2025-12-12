package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type RepoStatus struct {
	Path          string
	Name          string
	Branch        string
	Upstream      string
	Ahead         int
	Behind        int
	Dirty         bool
	HasUpstream   bool
	Error         error
	Fetching      bool
	Rebasing      bool
	Pushing       bool
	LastMessage   string
	CommitSubject string
	CommitAge     string
	CommitTime    int64 // Unix timestamp for sorting
}

func (s *RepoStatus) IsSynced() bool {
	return s.HasUpstream && s.Ahead == 0 && s.Behind == 0 && s.Error == nil
}

func (s *RepoStatus) NeedsPush() bool {
	return s.HasUpstream && s.Ahead > 0 && s.Error == nil
}

func (s *RepoStatus) NeedsPull() bool {
	return s.HasUpstream && s.Behind > 0 && s.Error == nil
}

func GetStatus(path, name string) *RepoStatus {
	status := &RepoStatus{
		Path: path,
		Name: name,
	}

	// Check if path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		status.Error = fmt.Errorf("path does not exist")
		return status
	}
	if err != nil {
		status.Error = fmt.Errorf("cannot access path")
		return status
	}
	if !info.IsDir() {
		status.Error = fmt.Errorf("not a directory")
		return status
	}

	// Check if it's a git repo
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		status.Error = fmt.Errorf("not a git repo")
		return status
	}

	// Get current branch
	branch, err := runGit(path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		status.Error = fmt.Errorf("no commits yet")
		return status
	}
	status.Branch = strings.TrimSpace(branch)

	// Check for uncommitted changes
	porcelain, _ := runGit(path, "status", "--porcelain")
	status.Dirty = strings.TrimSpace(porcelain) != ""

	// Get last commit info
	commitInfo, err := runGit(path, "log", "-1", "--format=%s|%cr|%ct")
	if err == nil {
		parts := strings.SplitN(strings.TrimSpace(commitInfo), "|", 3)
		if len(parts) >= 2 {
			status.CommitSubject = parts[0]
			status.CommitAge = parts[1]
		}
		if len(parts) == 3 {
			status.CommitTime, _ = strconv.ParseInt(parts[2], 10, 64)
		}
	}

	// Get upstream
	upstream, err := runGit(path, "rev-parse", "--abbrev-ref", "@{upstream}")
	if err != nil {
		status.HasUpstream = false
		return status
	}
	status.Upstream = strings.TrimSpace(upstream)
	status.HasUpstream = true

	// Get ahead/behind counts
	revList, err := runGit(path, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if err != nil {
		status.Error = fmt.Errorf("failed to get ahead/behind: %w", err)
		return status
	}

	parts := strings.Fields(strings.TrimSpace(revList))
	if len(parts) == 2 {
		status.Ahead, _ = strconv.Atoi(parts[0])
		status.Behind, _ = strconv.Atoi(parts[1])
	}

	return status
}

func Fetch(path string) error {
	_, err := runGit(path, "fetch", "--prune")
	return err
}

func Pull(path string) error {
	_, err := runGit(path, "pull", "--rebase", "--autostash")
	return err
}

func Push(path string) error {
	_, err := runGit(path, "push")
	return err
}

// Remote represents a git remote
type Remote struct {
	Name string
	URL  string
}

// ListRemotes returns all configured remotes for a repository
func ListRemotes(path string) ([]Remote, error) {
	output, err := runGit(path, "remote", "-v")
	if err != nil {
		return nil, err
	}

	remoteMap := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			// Only take fetch URLs (avoid duplicates from push)
			if len(parts) >= 3 && strings.Contains(parts[2], "fetch") {
				remoteMap[parts[0]] = parts[1]
			} else if _, exists := remoteMap[parts[0]]; !exists {
				remoteMap[parts[0]] = parts[1]
			}
		}
	}

	var remotes []Remote
	for name, url := range remoteMap {
		remotes = append(remotes, Remote{Name: name, URL: url})
	}

	// Sort by name for consistent ordering
	sort.Slice(remotes, func(i, j int) bool {
		// "origin" should come first
		if remotes[i].Name == "origin" {
			return true
		}
		if remotes[j].Name == "origin" {
			return false
		}
		return remotes[i].Name < remotes[j].Name
	})

	return remotes, nil
}

// RemoteBranch represents a branch on a remote
type RemoteBranch struct {
	Remote string
	Branch string
}

// ListRemoteBranches returns branches available on remotes that match the given branch name
func ListRemoteBranches(path, branchName string) ([]RemoteBranch, error) {
	// First fetch to ensure we have up-to-date remote info
	output, err := runGit(path, "branch", "-r")
	if err != nil {
		return nil, err
	}

	var branches []RemoteBranch
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "->") {
			continue
		}
		// Parse "origin/main" format
		parts := strings.SplitN(line, "/", 2)
		if len(parts) == 2 {
			remote := parts[0]
			branch := parts[1]
			// Match exact branch name or show all if branchName is empty
			if branchName == "" || branch == branchName {
				branches = append(branches, RemoteBranch{Remote: remote, Branch: branch})
			}
		}
	}

	return branches, nil
}

// SetUpstream sets the upstream branch for the current branch
func SetUpstream(path, remote, branch string) error {
	upstream := remote + "/" + branch
	_, err := runGit(path, "branch", "--set-upstream-to="+upstream)
	return err
}

// PushWithUpstream pushes the current branch and sets upstream tracking
func PushWithUpstream(path, remote, branch string) error {
	_, err := runGit(path, "push", "-u", remote, branch)
	return err
}

// AddRemote adds a new remote to the repository
func AddRemote(path, name, url string) error {
	_, err := runGit(path, "remote", "add", name, url)
	return err
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("%s", errMsg)
	}

	return stdout.String(), nil
}
