package version

var version string

type BuildMetadata struct {
	Version string
}

func Current() BuildMetadata {
	currentVersion := "dev"

	if version != "" {
		currentVersion = version
	}

	return BuildMetadata{
		Version: currentVersion,
	}
}
