package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type RepoStatus struct {
	Path         string
	Name         string
	Branch       string
	Upstream     string
	Ahead        int
	Behind       int
	HasUpstream  bool
	Error        error
	Fetching     bool
	Rebasing     bool
	Pushing      bool
	LastMessage  string
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

	// Get current branch
	branch, err := runGit(path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		status.Error = fmt.Errorf("not a git repo or no commits")
		return status
	}
	status.Branch = strings.TrimSpace(branch)

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
	_, err := runGit(path, "pull", "--rebase")
	return err
}

func Push(path string) error {
	_, err := runGit(path, "push")
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
