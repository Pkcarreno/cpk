package manifest_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/pkcarreno/cpk/internal/manifest"
)

func TestReadNotFound(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "package.json")

	_, err := manifest.Read(missingPath)
	if err == nil {
		t.Fatalf("expected an error, but got nil")
	}
	if !errors.Is(err, manifest.ErrManifestNotFound) {
		t.Fatalf("expected ErrManifestNotFound, got: %v", err)
	}
	expectedErrorMessage := "the package.json file does not exist"
	if err.Error() != expectedErrorMessage {
		t.Fatalf("expected %q, got %q", expectedErrorMessage, err.Error())
	}
}
