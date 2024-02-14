package version

import (
	"encoding/json"
	"fmt"
	"runtime"
)

var (
	GitCommit  string
	GitBranch  string
	GitSummary string
	BuildDate  string
	AppVersion string
	GoVersion  = runtime.Version()
)

type Version struct {
	GitCommit  string `json:"git_commit"`
	GitBranch  string `json:"git_branch"`
	GitSummary string `json:"git_summary"`
	BuildDate  string `json:"build_date"`
	AppVersion string `json:"app_version"`
	GoVersion  string `json:"go_version"`
}

func Current() *Version {
	return &Version{
		GitBranch:  GitBranch,
		GitCommit:  GitCommit,
		GitSummary: GitSummary,
		BuildDate:  BuildDate,
		AppVersion: AppVersion,
		GoVersion:  GoVersion,
	}
}

func (v *Version) String() string {
	return fmt.Sprintf("version=%s ref=%s branch=%s built=%s", v.AppVersion, v.GitCommit, v.GitBranch, v.BuildDate)
}

func (v *Version) MustJSON() json.RawMessage {
	byt, err := json.Marshal(v)
	if err != nil {
		panic("unable to marshal version")
	}
	return byt
}
