package mkversions

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"
)

// Опции для настройки Info
type Option func(*Info)

func WithCommitHash(hash string) Option {
	return func(info *Info) {
		info.GITInfo.CommitHash = hash
		if len(hash) > 7 {
			info.GITInfo.CommitHashShort = hash[:7]
		} else {
			info.GITInfo.CommitHashShort = hash
		}
	}
}

// С помощью функции-опции можно изменить только имя ветки
func WithBranchName(branch string) Option {
	return func(info *Info) {
		info.GITInfo.BranchName = branch
	}
}

// С помощью функции-опции можно изменить только дату коммита
func WithCommitDate(date time.Time) Option {
	return func(info *Info) {
		info.GITInfo.CommitDate = date
	}
}

// Опция для задания времени начала выборки логов (logSince)
func WithLogSince(since time.Time) Option {
	return func(info *Info) {
		info.GITInfo.ChangelogSince = since
	}
}

func runGitCommand(args ...string) (string, string, error) {
	fmt.Printf("git %v\n", args)
	cmd := exec.Command("git", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}
