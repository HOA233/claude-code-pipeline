package main

import (
	"fmt"
	"runtime"
)

// Build information (set via ldflags)
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// VersionInfo contains version information
type VersionInfo struct {
	Version    string `json:"version"`
	BuildTime  string `json:"build_time"`
	GitCommit  string `json:"git_commit"`
	GoVersion  string `json:"go_version"`
	Platform   string `json:"platform"`
}

// GetVersion returns the current version info
func GetVersion() VersionInfo {
	return VersionInfo{
		Version:    Version,
		BuildTime:  BuildTime,
		GitCommit:  GitCommit,
		GoVersion:  runtime.Version(),
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a formatted version string
func (v VersionInfo) String() string {
	return fmt.Sprintf("Claude Pipeline v%s (commit: %s, built: %s, %s)",
		v.Version, v.GitCommit, v.BuildTime, v.GoVersion)
}