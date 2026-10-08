package registry

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

type PackageMetadata struct {
	Name     string                    `json:"name"`
	DistTags map[string]string         `json:"dist-tags"`
	Versions map[string]PackageVersion `json:"versions"`
}

type PackageVersion struct {
	Dist PackageDist `json:"dist"`
}

type PackageDist struct {
	Integrity string `json:"integrity"`
}

var (
	ErrPackageNotFound       = errors.New("the package does not exist in the registry")
	ErrRegistryRequestFailed = errors.New("the registry request failed")
)

func FetchPackageMetadata(requestContext context.Context, httpClient *http.Client, registryURL, packageName string) (PackageMetadata, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	endpointURL, err := url.JoinPath(registryURL, packageName)
	if err != nil {
		return PackageMetadata{}, err
	}

	request, err := http.NewRequestWithContext(requestContext, http.MethodGet, endpointURL, nil)
	if err != nil {
		return PackageMetadata{}, err
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return PackageMetadata{}, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return PackageMetadata{}, ErrPackageNotFound
	}
	if response.StatusCode != http.StatusOK {
		return PackageMetadata{}, ErrRegistryRequestFailed
	}

	var packageMetadata PackageMetadata
	err = json.NewDecoder(response.Body).Decode(&packageMetadata)
	if err != nil {
		return PackageMetadata{}, err
	}

	return packageMetadata, nil
}
