package mkversions

import (
	"bytes"
	"os/exec"
	"time"
)

// Опции для настройки Info
type Option func(*Info)

func WithVersion(version string) Option {
	return func(info *Info) {
		info.Version = version
	}
}

func WithReleaseType(releaseType string) Option {
	return func(info *Info) {
		info.ReleaseType = releaseType
	}
}

func WithDeveloper(developer string) Option {
	return func(info *Info) {
		info.Developer = developer
	}
}

func WithProgramName(programName string) Option {
	return func(info *Info) {
		info.ProgramName = programName
	}
}

func WithCompanyName(companyName string) Option {
	return func(info *Info) {
		info.CompanyName = companyName
	}
}

func WithDescription(description string) Option {
	return func(info *Info) {
		info.Description = description
	}
}

func WithLegal(legal string) Option {
	return func(info *Info) {
		info.Legal = legal
	}
}

func WithProductVersion(productVersion string) Option {
	return func(info *Info) {
		info.ProductVersion = productVersion
	}
}

func WithGoVersion(goVersion string) Option {
	return func(info *Info) {
		info.GoVersion = goVersion
	}
}

func WithPlatform(platform string) Option {
	return func(info *Info) {
		info.Platform = platform
	}
}

func WithArchitecture(architecture string) Option {
	return func(info *Info) {
		info.Architecture = architecture
	}
}

func WithBuildID(buildID string) Option {
	return func(info *Info) {
		info.BuildID = buildID
	}
}

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

func WithBranchName(branch string) Option {
	return func(info *Info) {
		info.GITInfo.BranchName = branch
	}
}

func WithCommitDate(date time.Time) Option {
	return func(info *Info) {
		info.GITInfo.CommitDate = date
	}
}

func WithLogSince(since time.Time) Option {
	return func(info *Info) {
		info.GITInfo.ChangelogSince = since
	}
}

func runGitCommand(args ...string) (string, string, error) {
	cmd := exec.Command("git", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}
