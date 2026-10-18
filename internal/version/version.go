package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

var version string
var commit string
var date string
var readBuildInfo = debug.ReadBuildInfo

type BuildMetadata struct {
	Version  string
	Commit   string
	Date     string
	Go       string
	OS       string
	Arch     string
	Compiler string
}

func (buildMetadata BuildMetadata) String() string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "cpk %s\n", buildMetadata.Version)
	if buildMetadata.Commit != "" {
		fmt.Fprintf(&builder, "commit %s\n", buildMetadata.Commit)
	}
	if buildMetadata.Date != "" {
		fmt.Fprintf(&builder, "date %s\n", buildMetadata.Date)
	}
	fmt.Fprintf(&builder, "go %s %s/%s\n", buildMetadata.Go, buildMetadata.OS, buildMetadata.Arch)

	return builder.String()
}

func Current() BuildMetadata {
	build := BuildMetadata{
		Version:  version,
		Commit:   commit,
		Date:     date,
		Go:       runtime.Version(),
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Compiler: runtime.Compiler,
	}

	if info, ok := readBuildInfo(); ok {
		if build.Version == "" && info.Main.Version != "" && info.Main.Version != "(devel)" {
			build.Version = info.Main.Version
		}

		injected := commit != ""
		var modified bool

		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if build.Commit == "" {
					build.Commit = setting.Value
				}
			case "vcs.time":
				if build.Date == "" {
					build.Date = setting.Value
				}
			case "vcs.modified":
				modified = setting.Value == "true"
			}
		}

		if modified && !injected && build.Commit != "" {
			build.Commit += "-dirty"
		}
	}

	if build.Version == "" {
		build.Version = "dev"
	}
	build.Version = strings.TrimPrefix(build.Version, "v")
	build.Version = strings.TrimSuffix(build.Version, "+dirty")

	return build
}
