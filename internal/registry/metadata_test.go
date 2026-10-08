package registry_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkcarreno/cpk/internal/registry"
)

func TestFetchPackageMetadata(t *testing.T) {
	t.Run("fetches and parses package metadata successfully", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			if request.Method != http.MethodGet {
				responseWriter.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			if request.URL.Path != "/pnpm" {
				responseWriter.WriteHeader(http.StatusNotFound)
				return
			}

			responseWriter.Header().Set("Content-Type", "application/json")
			responseWriter.WriteHeader(http.StatusOK)
			responseWriter.Write([]byte(`{
				"name": "pnpm",
				"dist-tags": {
					"latest": "9.1.0"
				},
				"versions": {
					"9.1.0": {
						"dist": {
							"integrity": "sha512-mockhash"
						}
					}
				}
			}`))
		}))
		t.Cleanup(mockServer.Close)

		requestContext := context.Background()
		metadata, err := registry.FetchPackageMetadata(requestContext, mockServer.Client(), mockServer.URL, "pnpm")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if metadata.Name != "pnpm" {
			t.Fatalf("expected name %q, got: %q", "pnpm", metadata.Name)
		}

		if metadata.DistTags["latest"] != "9.1.0" {
			t.Fatalf("expected latest tag %q, got: %q", "9.1.0", metadata.DistTags["latest"])
		}

		if actualIntegrity := metadata.Versions["9.1.0"].Dist.Integrity; actualIntegrity != "sha512-mockhash" {
			t.Fatalf("expected integrity %q, got: %q", "sha512-mockhash", actualIntegrity)
		}
	})

	t.Run("returns ErrPackageNotFound when registry returns 404", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			if request.Method != http.MethodGet {
				responseWriter.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			if request.URL.Path != "/pnpm" {
				responseWriter.WriteHeader(http.StatusNotFound)
				return
			}

			responseWriter.WriteHeader(http.StatusNotFound)
		}))
		t.Cleanup(mockServer.Close)

		requestContext := context.Background()
		_, err := registry.FetchPackageMetadata(requestContext, mockServer.Client(), mockServer.URL, "pnpm")
		if !errors.Is(err, registry.ErrPackageNotFound) {
			t.Fatalf("expected ErrPackageNotFound, got: %v", err)
		}
	})

	t.Run("returns ErrRegistryRequestFailed when registry returns unexpected status", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			if request.Method != http.MethodGet {
				responseWriter.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			if request.URL.Path != "/pnpm" {
				responseWriter.WriteHeader(http.StatusNotFound)
				return
			}

			responseWriter.WriteHeader(http.StatusInternalServerError)
		}))
		t.Cleanup(mockServer.Close)

		requestContext := context.Background()
		_, err := registry.FetchPackageMetadata(requestContext, mockServer.Client(), mockServer.URL, "pnpm")
		if !errors.Is(err, registry.ErrRegistryRequestFailed) {
			t.Fatalf("expected ErrRegistryRequestFailed, got: %v", err)
		}
	})

	t.Run("returns error when context is canceled", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(mockServer.Close)

		canceledContext, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := registry.FetchPackageMetadata(canceledContext, mockServer.Client(), mockServer.URL, "pnpm")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}
