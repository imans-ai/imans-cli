package version

import "runtime"

var (
	Version       = "dev"
	Commit        = "unknown"
	BuildDate     = "unknown"
	SchemaVersion = "dev"
)

type Info struct {
	Version       string `json:"version"`
	Commit        string `json:"commit"`
	BuildDate     string `json:"build_date"`
	SchemaVersion string `json:"schema_version"`
	GOOS          string `json:"goos"`
	GOARCH        string `json:"goarch"`
}

func Current() Info {
	return Info{
		Version:       Version,
		Commit:        Commit,
		BuildDate:     BuildDate,
		SchemaVersion: SchemaVersion,
		GOOS:          runtime.GOOS,
		GOARCH:        runtime.GOARCH,
	}
}
