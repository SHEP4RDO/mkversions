package mkversions

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type GITInfo struct {
	CommitHash      string
	CommitHashShort string
	BranchName      string
	CommitDate      string
	*Changelog
}

type Changelog struct {
	Entries []string
	Commits []CommitDetails
}

type CommitDetails struct {
	Hash    string `json:"hash"`
	Message string `json:"message"`
	Author  string `json:"author"`
	Email   string `json:"email"`
	Date    string `json:"date"`
}

func GetGitCommitHashFull() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Git commit hash: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func GetGitCommitHashShort() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Git commit hash: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func GetGitBranchName() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Git branch name: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func GetGitCommitDate() (string, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%cd", "--date=short")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Git commit date: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func GetGitChangelog(since string) (*Changelog, error) {
	cmdArgs := []string{"log", "--pretty=format:%h - %s - %an <%ae> - %ad", "--no-merges", "--date=iso"}
	if since != "" {
		cmdArgs = append(cmdArgs, fmt.Sprintf("%s..HEAD", since))
	}
	cmd := exec.Command("git", cmdArgs...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get Git changelog: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	entries := make([]string, len(lines))
	commits := make([]CommitDetails, len(lines))

	for i, line := range lines {
		entries[i] = line

		parts := strings.SplitN(line, " - ", 4)
		if len(parts) < 4 {
			continue
		}

		authorPart := parts[2]
		emailStart := strings.Index(authorPart, "<")
		emailEnd := strings.Index(authorPart, ">")
		email := authorPart[emailStart+1 : emailEnd]

		commits[i] = CommitDetails{
			Hash:    parts[0],
			Message: parts[1],
			Author:  strings.TrimSpace(authorPart[:emailStart-1]),
			Email:   email,
			Date:    parts[3],
		}
	}

	return &Changelog{Entries: entries, Commits: commits}, nil
}

func (cl *Changelog) ToMarkdown() string {
	var sb strings.Builder
	sb.WriteString("## Changelog\n\n")
	for _, entry := range cl.Entries {
		sb.WriteString(fmt.Sprintf("- %s\n", entry))
	}
	return sb.String()
}

func (cl *Changelog) ToJSON() (string, error) {
	data, err := json.Marshal(cl.Commits)
	if err != nil {
		return "", fmt.Errorf("failed to marshal changelog to JSON: %v", err)
	}
	return string(data), nil
}
