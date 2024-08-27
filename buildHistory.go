package mkversions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type BuildHistory struct {
	Builds []*Info
	Limit  int
}

func NewBuildHistory(limit int) *BuildHistory {
	return &BuildHistory{
		Builds: make([]*Info, 0, limit),
		Limit:  limit,
	}
}

func (bh *BuildHistory) AddBuild(info *Info) {
	if len(bh.Builds) >= bh.Limit {
		bh.Builds = bh.Builds[1:]
	}
	bh.Builds = append(bh.Builds, info)
}

func (bh *BuildHistory) GetLatestBuild() *Info {
	if len(bh.Builds) == 0 {
		return nil
	}
	return bh.Builds[len(bh.Builds)-1]
}

func (bh *BuildHistory) GetBuildByIndex(index int) (*Info, error) {
	if index < 0 || index >= len(bh.Builds) {
		return nil, fmt.Errorf("invalid index")
	}
	return bh.Builds[index], nil
}

func (bh *BuildHistory) ListBuilds() []*Info {
	return bh.Builds
}

func (bh *BuildHistory) SaveToFile(filePath string) error {
	data, err := json.Marshal(bh)
	if err != nil {
		return fmt.Errorf("failed to marshal build history: %v", err)
	}

	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write build history to file: %v", err)
	}

	return nil
}

func LoadBuildHistoryFromFile(filePath string) (*BuildHistory, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read build history file: %v", err)
	}

	var bh BuildHistory
	err = json.Unmarshal(data, &bh)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal build history: %v", err)
	}

	return &bh, nil
}
