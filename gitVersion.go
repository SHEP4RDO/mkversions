package mkversions

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type GITInfo struct {
	CommitHash      string
	CommitHashShort string
	BranchName      string
	CommitDate      time.Time
	ChangelogSince  time.Time
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

func (info *Info) PrepareGit() {
	if info.GITInfo.BranchName == "unknown" {
		var branchErr error
		info.GITInfo.BranchName, branchErr = GetGitBranchName()
		if branchErr != nil {
			fmt.Println("Error while getting git branch name: ", branchErr)
			info.GITInfo.BranchName = ""
		}
	}

	if info.GITInfo.CommitHash == "unknown" {
		var hashErr error
		info.GITInfo.CommitHash, hashErr = GetGitCommitHashFull(info.GITInfo.BranchName)
		if hashErr != nil {
			fmt.Println("Error while getting git commit hash: ", hashErr)
			info.GITInfo.CommitHash = "unknown"
			info.GITInfo.CommitHashShort = "unknown"
		}

		if len(info.GITInfo.CommitHash) > 7 {
			info.GITInfo.CommitHashShort = info.GITInfo.CommitHash[:7]
		} else {
			info.GITInfo.CommitHashShort = info.GITInfo.CommitHash
		}
	}

	if info.GITInfo.CommitDate.IsZero() {
		var dateErr error
		info.GITInfo.CommitDate, dateErr = GetGitCommitDate(info.GITInfo.BranchName)
		if dateErr != nil {
			fmt.Println("Error while getting git commit date: ", dateErr)
			info.GITInfo.CommitDate = time.Time{}
		}
	}

	var logSince time.Time
	if !info.ChangelogSince.IsZero() {
		logSince = info.GITInfo.CommitDate.Add(-24 * time.Hour)
	} else {
		logSince = time.Now().Add(-24 * time.Hour)
	}

	var changelogErr error
	info.GITInfo.Changelog, changelogErr = GetGitChangelog(logSince.Format("2006-01-02"), info.GITInfo.BranchName)
	if changelogErr != nil {
		fmt.Println("Error while getting git changelog: ", changelogErr)
		info.GITInfo.Changelog = &Changelog{}
	}
}

func GetGitCommitHashFull(ref string) (string, error) {
	args := []string{"rev-parse", "HEAD"}
	if ref != "" {
		args = []string{"rev-parse", ref}
	}

	stdout, stderr, err := runGitCommand(args...)
	if err != nil {
		return "", fmt.Errorf("failed to get Git commit hash: %v, %s", err, stderr)
	}
	return strings.TrimSpace(stdout), nil
}

func GetGitCommitHashShort(ref string) (string, error) {
	args := []string{"rev-parse", "--short", "HEAD"}
	if ref != "" {
		args = []string{"rev-parse", "--short", ref}
	}

	stdout, stderr, err := runGitCommand(args...)
	if err != nil {
		return "", fmt.Errorf("failed to get Git commit hash: %v, %s", err, stderr)
	}
	return strings.TrimSpace(stdout), nil
}

func GetGitBranchName() (string, error) {
	stdout, stderr, err := runGitCommand("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get Git branch name: %v, %s", err, stderr)
	}
	return strings.TrimSpace(stdout), nil
}

func GetGitCommitDate(ref string) (time.Time, error) {
	args := []string{"log", "-1", "--format=%cd", "--date=local"}
	if ref != "" {
		args = append(args, ref)
	}

	stdout, stderr, err := runGitCommand(args...)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get Git commit date: %v, %s", err, stderr)
	}

	strDate := strings.TrimSpace(stdout)
	date, err := time.Parse("Mon Jan 2 15:04:05 2006", strDate)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse Git commit date: %v", err)
	}
	return date, nil
}

// GetGitChangelog получает журнал коммитов Git с учетом даты и ссылки
func GetGitChangelog(since, ref string) (*Changelog, error) {
	var cmdArgs []string

	// Формирование аргументов команды
	cmdArgs = []string{"log", "--pretty=format:%h - %s - %an <%ae> - %ad", "--no-merges", "--date=iso"}

	// Если указана дата, добавляем аргумент --since
	if since != "" {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--since=%s", since))
	}

	// Если указана ссылка, добавляем ее
	if ref != "" {
		cmdArgs = append(cmdArgs, ref)
	}

	// Выполнение команды
	output, stderr, err := runGitCommand(cmdArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get Git changelog: %v, stderr: %s", err, stderr)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
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
