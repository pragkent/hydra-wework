package main

import (
	"fmt"
	"runtime"
)

// The following fields are populated at buildtime with bazel's linkstamp
// feature. This is equivalent to using golang directly with -ldflags -X.
var (
	buildVersion   string
	buildTime      string
	buildGitCommit string
	buildGitBranch string
)

// BuildInfo describes version information about the binary build.
type BuildInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"commit"`
	GitBranch string `json:"branch"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
}

func (b *BuildInfo) String() string {
	return fmt.Sprintf("%v-%v-%v", b.Version, b.GitBranch, b.GitCommit)
}

var (
	// Info exports the build version information.
	Info BuildInfo
)

func init() {
	Info.Version = buildVersion
	Info.GitCommit = buildGitCommit
	Info.GitBranch = buildGitBranch
	Info.BuildTime = buildTime
	Info.GoVersion = runtime.Version()
}

// Version returns a multi-line version information
func Version() string {
	return fmt.Sprintf(`Version: %v
GitCommit: %v
GitBranch: %v
GoVersion: %v
BuildTime: %v
`,
		Info.Version,
		Info.GitCommit,
		Info.GitBranch,
		Info.GoVersion,
		Info.BuildTime)
}
