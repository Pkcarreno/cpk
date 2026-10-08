package registry

import (
	"os"
)

const (
	defaultRegistryURL = "https://registry.npmjs.org"
)

func ResolveRegistryURL(customURL string) string {
	if customURL != "" {
		return customURL
	}

	environmentRegistryURL := os.Getenv("NPM_CONFIG_REGISTRY")
	if environmentRegistryURL != "" {
		return environmentRegistryURL
	}

	return defaultRegistryURL
}
