package resolver

import (
	"errors"
	"strconv"
	"strings"

	"github.com/pkcarreno/cpk/internal/registry"
)

var (
	ErrVersionNotFound = errors.New("the requested version does not exist")
)

type semanticVersion struct {
	raw          string
	major        int
	minor        int
	patch        int
	isPrerelease bool
}

func parseSemanticVersion(versionString string) (semanticVersion, bool) {
	mainVersion, _, _ := strings.Cut(versionString, "+")
	coreVersion, _, hasPrerelease := strings.Cut(mainVersion, "-")

	parts := strings.Split(coreVersion, ".")
	if len(parts) != 3 {
		return semanticVersion{}, false
	}

	major, errMajor := strconv.Atoi(parts[0])
	minor, errMinor := strconv.Atoi(parts[1])
	patch, errPatch := strconv.Atoi(parts[2])
	if errMajor != nil || errMinor != nil || errPatch != nil {
		return semanticVersion{}, false
	}

	return semanticVersion{
		raw:          versionString,
		major:        major,
		minor:        minor,
		patch:        patch,
		isPrerelease: hasPrerelease,
	}, true
}

func isHigherVersion(candidate, current semanticVersion) bool {
	if candidate.major != current.major {
		return candidate.major > current.major
	}
	if candidate.minor != current.minor {
		return candidate.minor > current.minor
	}
	return candidate.patch > current.patch
}

func ResolveVersion(metadata registry.PackageMetadata, query string) (string, error) {
	if version, hasTag := metadata.DistTags[query]; hasTag {
		if _, hasVersion := metadata.Versions[version]; hasVersion {
			return version, nil
		}
		return "", ErrVersionNotFound
	}

	if _, hasVersion := metadata.Versions[query]; hasVersion {
		return query, nil
	}

	if resolvedVersion, isFound := resolveNumericPrefix(metadata, query); isFound {
		return resolvedVersion, nil
	}

	return "", ErrVersionNotFound
}

func resolveNumericPrefix(metadata registry.PackageMetadata, query string) (string, bool) {
	parts := strings.Split(query, ".")
	if len(parts) > 2 {
		return "", false
	}

	majorPrefix, errMajor := strconv.Atoi(parts[0])
	if errMajor != nil || majorPrefix < 0 {
		return "", false
	}

	minorPrefix := -1
	if len(parts) == 2 {
		parsedMinor, errMinor := strconv.Atoi(parts[1])
		if errMinor != nil || parsedMinor < 0 {
			return "", false
		}
		minorPrefix = parsedMinor
	}

	var highestVersion semanticVersion
	hasMatch := false

	for rawVersion := range metadata.Versions {
		parsedVersion, isParsed := parseSemanticVersion(rawVersion)
		if !isParsed || parsedVersion.isPrerelease {
			continue
		}

		if parsedVersion.major != majorPrefix {
			continue
		}

		if minorPrefix != -1 && parsedVersion.minor != minorPrefix {
			continue
		}

		if !hasMatch || isHigherVersion(parsedVersion, highestVersion) {
			highestVersion = parsedVersion
			hasMatch = true
		}
	}

	if !hasMatch {
		return "", false
	}

	return highestVersion.raw, true
}

func IsLowerVersion(candidateVersion, currentVersion string) (bool, error) {
	candidate, isCandidateParsed := parseSemanticVersion(candidateVersion)
	if !isCandidateParsed {
		return false, ErrVersionNotFound
	}
	current, isCurrentParsed := parseSemanticVersion(currentVersion)
	if !isCurrentParsed {
		return false, ErrVersionNotFound
	}
	return isHigherVersion(current, candidate), nil
}

func ResolvePackageName(packageManager, targetVersion string) string {
	if packageManager != "yarn" {
		return packageManager
	}

	if targetVersion == "" || targetVersion == "latest" {
		return "@yarnpkg/cli-dist"
	}

	cleanVersion, _, _ := strings.Cut(targetVersion, "-")
	cleanVersion, _, _ = strings.Cut(cleanVersion, "+")
	parts := strings.Split(cleanVersion, ".")

	if major, err := strconv.Atoi(parts[0]); err == nil && major >= 0 && major < 2 {
		return "yarn"
	}

	return "@yarnpkg/cli-dist"
}
