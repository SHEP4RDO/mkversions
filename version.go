package mkversions

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Info хранит информацию о версии сборки
type Info struct {
	Version         string
	BuildDate       time.Time
	GoVersion       string
	Platform        string
	BuildID         string
	ReleaseType     string
	Architecture    string
	Developer       string
	Dependencies    map[string]string
	DetailedVersion string
	*GITInfo
	*AppMetadata
}

// Функция создания Info
func NewInfo(version, releaseType, developer string, opts ...Option) *Info {
	var branchName string
	var commitHash string
	var commitDate time.Time

	// Установка значений по умолчанию
	branchName = "unknown"
	commitHash = "unknown"
	commitDate = time.Time{}

	// Создание начального объекта Info с дефолтными значениями
	info := &Info{
		Version:         version,
		BuildDate:       time.Now(),
		GoVersion:       runtime.Version(),
		Platform:        runtime.GOOS,
		Architecture:    runtime.GOARCH,
		BuildID:         generateBuildID(),
		ReleaseType:     releaseType,
		Dependencies:    make(map[string]string),
		Developer:       developer,
		DetailedVersion: fmt.Sprintf("%s-%s(%s) | %s", version, commitHash, releaseType, time.Now().Format("2006-01-02")),
		GITInfo: &GITInfo{
			CommitHash:      commitHash,
			CommitHashShort: commitHash[:7],
			BranchName:      branchName,
			CommitDate:      commitDate,
			Changelog:       nil,
		},
		AppMetadata: &AppMetadata{},
	}

	// Применение опций
	for _, opt := range opts {
		opt(info)
	}

	info.PrepareGit()
	return info
}

func (i *Info) SetInfo(opts ...Option) *Info {
	for _, opt := range opts {
		opt(i)
	}

	return i
}

// NewInfo создает новый объект Info с заданной версией и коммитом
func NewInfoCustom(version, commit, commitFull, releaseType, developer, branch string, commitDate time.Time) *Info {
	dep, err := getDependencies()
	if err != nil {
		fmt.Println("Error while getting dependencies: ", err)
	}

	buildDate := time.Now().Format("2006-01-02")
	return &Info{
		Version:         version,
		BuildDate:       time.Now(),
		GoVersion:       runtime.Version(),
		Platform:        runtime.GOOS,
		Architecture:    runtime.GOARCH,
		BuildID:         generateBuildID(),
		ReleaseType:     releaseType,
		Dependencies:    dep,
		Developer:       developer,
		DetailedVersion: fmt.Sprintf("%s-%s(%s) | %s", version, commit, releaseType, buildDate),
		GITInfo: &GITInfo{
			CommitHash:      commitFull,
			CommitHashShort: commit,
			BranchName:      branch,
			CommitDate:      commitDate,
		},
	}
}

// generateBuildID создает уникальный идентификатор для сборки
func generateBuildID() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		panic("failed to generate build ID")
	}
	return hex.EncodeToString(bytes)
}

func getDependencies() (map[string]string, error) {
	cmd := exec.Command("go", "list", "-m", "-json", "all")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var modules []struct {
		Path    string `json:"Path"`
		Version string `json:"Version"`
	}

	decoder := json.NewDecoder(strings.NewReader(string(output)))
	for {
		var module struct {
			Path    string `json:"Path"`
			Version string `json:"Version"`
		}
		if err := decoder.Decode(&module); err != nil {
			break
		}
		modules = append(modules, module)
	}

	dependencies := make(map[string]string)
	for _, module := range modules {
		dependencies[module.Path] = module.Version
	}

	return dependencies, nil
}

func (info *Info) AddDependency(modulePath, version string) {
	if info.Dependencies == nil {
		info.Dependencies = make(map[string]string)
	}
	info.Dependencies[modulePath] = version
}

func (info *Info) RemoveDependency(modulePath string) {
	delete(info.Dependencies, modulePath)
}

func (info *Info) SaveToHistory(bh *BuildHistory) {
	bh.AddBuild(info)
}

// String возвращает информацию о версии в формате строки
func (info *Info) String() string {
	return fmt.Sprintf(
		"Version: %s\nBuild Date: %s\nCommit: %s\nGo Version: %s\nPlatform: %s\nArchitecture: %s\nBuild ID: %s\nRelease Type: %s\nDeveloper: %s\nDetailed Version: %s\nDependencies: %v",
		info.Version, info.BuildDate, info.CommitHash, info.GoVersion, info.Platform, info.Architecture, info.BuildID, info.ReleaseType, info.Developer, info.DetailedVersion, info.Dependencies,
	)
}

func (info *Info) ToMarkdown() string {
	return fmt.Sprintf(
		"## Version Info\n\n"+
			"* **Version:** %s\n"+
			"* **Build Date:** %s\n"+
			"* **Commit Hash:** %s\n"+
			"* **Go Version:** %s\n"+
			"* **Platform:** %s\n"+
			"* **Architecture:** %s\n"+
			"* **Build ID:** %s\n"+
			"* **Release Type:** %s\n"+
			"* **Developer:** %s\n"+
			"* **Detailed Version:** %s\n"+
			"* **Dependencies:** %v",
		info.Version, info.BuildDate, info.CommitHash, info.GoVersion, info.Platform, info.Architecture, info.BuildID, info.ReleaseType, info.Developer, info.DetailedVersion, info.Dependencies,
	)
}

// JSON возвращает информацию о версии в формате JSON
func (info *Info) JSON() string {
	data, err := json.Marshal(info)
	if err != nil {
		log.Fatalf("failed to marshal version info to JSON: %v", err)
	}
	return string(data)
}
