package manifest_test

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"path/filepath"
	"testing"

	"github.com/pkcarreno/cpk/internal/manifest"
)

func TestPackageSpec_String(t *testing.T) {
	t.Run("formats spec with integrity hash", func(t *testing.T) {
		spec := manifest.PackageSpec{
			Name:    "pnpm",
			Version: "9.2.0",
			Hash:    "+sha512.abcdef",
		}
		expected := "pnpm@9.2.0+sha512.abcdef"
		if spec.String() != expected {
			t.Fatalf("expected: %q, got: %q", expected, spec.String())
		}
	})

	t.Run("formats spec without integrity hash", func(t *testing.T) {
		spec := manifest.PackageSpec{
			Name:    "yarn",
			Version: "4.1.0",
		}
		expected := "yarn@4.1.0"
		if spec.String() != expected {
			t.Fatalf("expected: %q, got: %q", expected, spec.String())
		}
	})
}

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

func TestParsePackageSpec(t *testing.T) {
	sha1Hex := strings.Repeat("a", 40)
	sha224Hex := strings.Repeat("b", 56)
	sha256Hex := strings.Repeat("c", 64)
	sha384Hex := strings.Repeat("d", 96)
	sha512Hex := strings.Repeat("e", 128)

	validCases := []struct {
		name         string
		rawSpec      string
		expectedSpec manifest.PackageSpec
	}{
        {
            name:         "parses a valid package specification without an integrity hash",
            rawSpec:      "yarn@4.1.0",
            expectedSpec: manifest.PackageSpec{
                Name:    "yarn",
                Version: "4.1.0",
                Hash:    "",
            },
        },
        {
            name:         "parses a valid package specification with sha1 integrity hash",
            rawSpec:      "pnpm@9.1.0+sha1." + sha1Hex,
            expectedSpec: manifest.PackageSpec{
                Name:    "pnpm",
                Version: "9.1.0",
                Hash:    "+sha1." + sha1Hex,
            },
        },
        {
            name:         "parses a valid package specification with sha224 integrity hash",
            rawSpec:      "yarn@4.1.0+sha224." + sha224Hex,
            expectedSpec: manifest.PackageSpec{
                Name:    "yarn",
                Version: "4.1.0",
                Hash:    "+sha224." + sha224Hex,
            },
        },
        {
            name:         "parses a valid package specification with sha256 integrity hash",
            rawSpec:      "pnpm@9.1.0+sha256." + sha256Hex,
            expectedSpec: manifest.PackageSpec{
                Name:    "pnpm",
                Version: "9.1.0",
                Hash:    "+sha256." + sha256Hex,
            },
        },
        {
            name:         "parses a valid package specification with sha384 integrity hash",
            rawSpec:      "npm@10.2.0+sha384." + sha384Hex,
            expectedSpec: manifest.PackageSpec{
                Name:    "npm",
                Version: "10.2.0",
                Hash:    "+sha384." + sha384Hex,
            },
        },
        {
            name:         "parses a valid package specification with sha512 integrity hash",
            rawSpec:      "pnpm@9.1.0+sha512." + sha512Hex,
            expectedSpec: manifest.PackageSpec{
                Name:    "pnpm",
                Version: "9.1.0",
                Hash:    "+sha512." + sha512Hex,
            },
        },
	}

	for _, testCase := range validCases {
		t.Run(testCase.name, func(t *testing.T) {
			spec, err := manifest.ParsePackageSpec(testCase.rawSpec)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if spec != testCase.expectedSpec {
				t.Fatalf("expected spec %+v, got %+v", testCase.expectedSpec, spec)
			}
		})
	}

	invalidCases := []struct {
		name        string
		rawSpec     string
		expectedErr error
	}{
		{
            name:        "returns ErrInvalidPackageSpec on invalid format",
            rawSpec:     "invalid-spec",
            expectedErr: manifest.ErrInvalidPackageSpec,
        },
        {
            name:        "returns ErrInvalidPackageSpec on empty hash delimiter",
            rawSpec:     "pnpm@9.1.0+",
            expectedErr: manifest.ErrInvalidPackageSpec,
        },
        {
            name:        "returns ErrInvalidIntegrityHash when the algorithm is unsupported",
            rawSpec:     "pnpm@9.1.0+md5." + strings.Repeat("a", 32),
            expectedErr: manifest.ErrInvalidIntegrityHash,
        },
        {
            name:        "returns ErrInvalidIntegrityHash when the hash lacks dot separator",
            rawSpec:     "pnpm@9.1.0+sha512" + sha512Hex,
            expectedErr: manifest.ErrInvalidIntegrityHash,
        },
        {
            name:        "returns ErrInvalidIntegrityHash when the sha1 length is short",
            rawSpec:     "pnpm@9.1.0+sha1." + strings.Repeat("a", 39),
            expectedErr: manifest.ErrInvalidIntegrityHash,
        },
        {
            name:        "returns ErrInvalidIntegrityHash when the sha1 length is long",
            rawSpec:     "pnpm@9.1.0+sha1." + strings.Repeat("a", 41),
            expectedErr: manifest.ErrInvalidIntegrityHash,
        },
        {
            name:        "returns ErrInvalidIntegrityHash when the sha224 length is short",
            rawSpec:     "yarn@4.1.0+sha224." + strings.Repeat("a", 55),
            expectedErr: manifest.ErrInvalidIntegrityHash,
        },
        {
            name:        "returns ErrInvalidIntegrityHash when the sha256 length is short",
            rawSpec:     "pnpm@9.1.0+sha256." + strings.Repeat("a", 63),
            expectedErr: manifest.ErrInvalidIntegrityHash,
        },
        {
            name:        "returns ErrInvalidIntegrityHash when the sha384 length is short",
            rawSpec:     "npm@10.2.0+sha384." + strings.Repeat("a", 95),
            expectedErr: manifest.ErrInvalidIntegrityHash,
        },
        {
            name:        "returns ErrInvalidIntegrityHash when the sha512 length is short",
            rawSpec:     "pnpm@9.1.0+sha512." + strings.Repeat("a", 127),
            expectedErr: manifest.ErrInvalidIntegrityHash,
        },
        {
            name:        "returns ErrInvalidIntegrityHash when the sha512 length is long",
            rawSpec:     "pnpm@9.1.0+sha512." + strings.Repeat("a", 129),
            expectedErr: manifest.ErrInvalidIntegrityHash,
        },
        {
            name:        "returns ErrInvalidIntegrityHash when the hash contains non-hexadecimal characters",
            rawSpec:     "pnpm@9.1.0+sha1." + strings.Repeat("g", 40),
            expectedErr: manifest.ErrInvalidIntegrityHash,
        },
        {
            name:        "returns ErrInvalidIntegrityHash when the hash contains uppercase characters",
            rawSpec:     "pnpm@9.1.0+sha1." + strings.Repeat("A", 40),
            expectedErr: manifest.ErrInvalidIntegrityHash,
        },     	  	
   	}

	for _, testCase := range invalidCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := manifest.ParsePackageSpec(testCase.rawSpec)
			if !errors.Is(err, testCase.expectedErr) {
				t.Fatalf("expected error %v, got %v", testCase.expectedErr, err)
			}
		})
	}
}

func TestReadPackageSpec(t *testing.T) {
	t.Run("returns ErrManifestNotFound when file does not exist", func(t *testing.T) {
		missingPath := filepath.Join(t.TempDir(), "package.json")
		_, err := manifest.ReadPackageSpec(missingPath)
		if !errors.Is(err, manifest.ErrManifestNotFound) {
			t.Fatalf("expected ErrManifestNotFound, got: %v", err)
		}
	})

	t.Run("returns ErrMissingPackageManager when key is missing", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "package.json")
		err := os.WriteFile(path, []byte(`{"name": "test-pkg"}`), 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err = manifest.ReadPackageSpec(path)
		if !errors.Is(err, manifest.ErrMissingPackageManager) {
			t.Fatalf("expected ErrMissingPackageManager, got: %v", err)
		}
	})

	t.Run("returns parsed spec when key exists", func(t *testing.T) {
		validHash := "+sha512." + strings.Repeat("a", 128)
		path := filepath.Join(t.TempDir(), "package.json")
		err := os.WriteFile(path, []byte(`{"packageManager": "pnpm@9.1.0`+validHash+`"}`), 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		spec, err := manifest.ReadPackageSpec(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if spec.Name != "pnpm" || spec.Version != "9.1.0" || spec.Hash != validHash {
			t.Fatalf("unexpected spec: %+v", spec)
		}
	})

	t.Run("returns error when manifest contains invalid JSON", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "package.json")
		err := os.WriteFile(path, []byte(`{invalid-json}`), 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err = manifest.ReadPackageSpec(path)
		if !errors.Is(err, manifest.ErrInvalidManifest) {
			t.Fatalf("expected ErrInvalidiManifest, got: %v", err)
		}
	})
}

func TestUpdatePackageSpec(t *testing.T) {

	testCases := []struct {
		name    string
		fixture string
	}{
		{name: "preserves 2 spaces indentation", fixture: "existing_2spaces"},
		{name: "preserves 4 spaces indentation", fixture: "existing_4spaces"},
		{name: "preserves tab indentation", fixture: "existing_tabs"},
		{name: "preserves unformatted JSON", fixture: "existing_unformatted"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			inputBytes, err := os.ReadFile(filepath.Join("testdata", testCase.fixture+".json"))
			if err != nil {
				t.Fatalf("failed to read input test fixture: %v", err)
			}

			goldenBytes, err := os.ReadFile(filepath.Join("testdata", testCase.fixture+".golden.json"))
			if err != nil {
				t.Fatalf("failed to read golden test fixture: %v", err)
			}

			targetSpec := manifest.PackageSpec{
				Name:    "pnpm",
				Version: "9.2.0",
				Hash:    "+sha512.abcdef",
			}

			updatedBytes, err := manifest.UpdatePackageSpec(inputBytes, targetSpec)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !bytes.Equal(updatedBytes, goldenBytes) {
				t.Fatalf("output does not match golden file:\nexpected:\n%s\ngot:\n%s", string(goldenBytes), string(updatedBytes))
			}
		})
	}

	t.Run("returns error when manifest contains invalid JSON", func(t *testing.T) {
		invalidBytes := []byte(`{invalid-json}`)
		targetSpec := manifest.PackageSpec{
			Name:    "pnpm",
			Version: "9.2.0",
			Hash:    "+sha512.abcdef",
		}

		_, err := manifest.UpdatePackageSpec(invalidBytes, targetSpec)
		if !errors.Is(err, manifest.ErrInvalidManifest) {
			t.Fatalf("expected ErrInvalidManifest, got: %v", err)
		}
	})

	t.Run("returns ErrMissingPackageManager when key is absent", func(t *testing.T) {
		manifestBytes := []byte(`{"name": "test-pkg"}`)
		targetSpec := manifest.PackageSpec{
			Name:    "pnpm",
			Version: "9.2.0",
		}

		_, err := manifest.UpdatePackageSpec(manifestBytes, targetSpec)
		if !errors.Is(err, manifest.ErrMissingPackageManager) {
			t.Fatalf("expected ErrMissingPackageManager, got %v", err)
		}
	})
}

func TestInsertPackageSpec(t *testing.T) {

	testCases := []struct {
		name    string
		fixture string
	}{
		{name: "preserves 2 spaces indentation", fixture: "insert_2spaces"},
		{name: "preserves 4 spaces indentation", fixture: "insert_4spaces"},
		{name: "preserves tab indentation", fixture: "insert_tabs"},
		{name: "preserves unformatted JSON", fixture: "insert_unformatted"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			inputBytes, err := os.ReadFile(filepath.Join("testdata", testCase.fixture+".json"))
			if err != nil {
				t.Fatalf("failed to read input test fixture: %v", err)
			}

			goldenBytes, err := os.ReadFile(filepath.Join("testdata", testCase.fixture+".golden.json"))
			if err != nil {
				t.Fatalf("failed to read golden test fixture: %v", err)
			}

			targetSpec := manifest.PackageSpec{
				Name:    "pnpm",
				Version: "9.2.0",
				Hash:    "+sha512.abcdef",
			}

			updatedBytes, err := manifest.InsertPackageSpec(inputBytes, targetSpec)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !bytes.Equal(updatedBytes, goldenBytes) {
				t.Fatalf("output does not match golden file:\nexpected:\n%s\ngot:\n%s", string(goldenBytes), string(updatedBytes))
			}
		})
	}

	t.Run("returns error when manifest contains invalid JSON", func(t *testing.T) {
		invalidBytes := []byte(`{invalid-json}`)
		targetSpec := manifest.PackageSpec{
			Name:    "pnpm",
			Version: "9.2.0",
			Hash:    "+sha512.abcdef",
		}

		_, err := manifest.InsertPackageSpec(invalidBytes, targetSpec)
		if !errors.Is(err, manifest.ErrInvalidManifest) {
			t.Fatalf("expected ErrInvalidManifest, got: %v", err)
		}
	})

	t.Run("returns ErrPackageManagerExists when packageManager exists", func(t *testing.T) {
		manifestBytes := []byte(`{"packageManager": "pnpm@9.2.0"}`)
		targetSpec := manifest.PackageSpec{
			Name:    "pnpm",
			Version: "9.2.0",
		}

		_, err := manifest.InsertPackageSpec(manifestBytes, targetSpec)
		if !errors.Is(err, manifest.ErrPackageManagerExists) {
			t.Fatalf("expected ErrPackageManagerExists, got %v", err)
		}
	})
}

func TestWritePackageSpec(t *testing.T) {
	t.Run("leaves file untouched when spec already matches", func(t *testing.T) {
		validHash := "+sha512." + strings.Repeat("a", 128)
		manifestPath := filepath.Join(t.TempDir(), "package.json")
		rawContent := `{"name": "test-pkg", "packageManager":"pnpm@9.1.0`+validHash+`"}`
		err := os.WriteFile(manifestPath, []byte(rawContent), 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		initialStat, err := os.Stat(manifestPath)
		if err != nil {
			t.Fatalf("failed to stat file: %v", err)
		}

		matchingSpec := manifest.PackageSpec{
			Name:    "pnpm",
			Version: "9.1.0",
			Hash:    validHash,
		}

		isModified, err := manifest.WritePackageSpec(manifestPath, matchingSpec)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if isModified {
			t.Fatalf("expected isModified to be false, got true")
		}

		finalStat, err := os.Stat(manifestPath)
		if err != nil {
			t.Fatalf("failed to stat file: %v", err)
		}

		if !finalStat.ModTime().Equal(initialStat.ModTime()) {
			t.Fatalf("expected file modification time to remain unchanged")
		}
	})

	t.Run("updates file when spec differs", func(t *testing.T) {
		manifestPath := filepath.Join(t.TempDir(), "package.json")
		initialHash := "+sha512." + strings.Repeat("a", 128)
		rawContent := `{"name": "test-pkg", "packageManager":"pnpm@9.1.0`+initialHash+`"}`
		err := os.WriteFile(manifestPath, []byte(rawContent), 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		newHash := "+sha512." + strings.Repeat("b", 128)
		differentSpec := manifest.PackageSpec{
			Name:    "pnpm",
			Version: "9.2.0",
			Hash:    newHash,
		}

		isModified, err := manifest.WritePackageSpec(manifestPath, differentSpec)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !isModified {
			t.Fatalf("expected isModified to be true, got false")
		}

		currentSpec, err := manifest.ReadPackageSpec(manifestPath)
		if err != nil {
			t.Fatalf("unexpected error reading package spec: %v", err)
		}

		if currentSpec.String() != differentSpec.String() {
			t.Fatalf("current spec does not match target spec:\nexpected:\n%s\ngot:\n%s", currentSpec.String(), differentSpec.String())
		}
	})

	t.Run("inserts packageManager field when absent", func(t *testing.T) {
		manifestPath := filepath.Join(t.TempDir(), "package.json")
		rawContent := `{"name": "test-pkg"}`
		err := os.WriteFile(manifestPath, []byte(rawContent), 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		targetSpec := manifest.PackageSpec{
			Name:    "pnpm",
			Version: "9.2.0",
			Hash:    "+sha512." + strings.Repeat("a", 128),
		}

		isModified, err := manifest.WritePackageSpec(manifestPath, targetSpec)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !isModified {
			t.Fatalf("expected isModified to be true, got false")
		}

		currentSpec, err := manifest.ReadPackageSpec(manifestPath)
		if err != nil {
			t.Fatalf("unexpected error reading package spec: %v", err)
		}

		if currentSpec.String() != targetSpec.String() {
			t.Fatalf("current spec does not match target spec:\nexpected:\n%s\ngot:\n%s", currentSpec.String(), targetSpec.String())
		}
	})

	t.Run("returns error when manifest contains invalid JSON", func(t *testing.T) {
		invalidManifestPath := filepath.Join(t.TempDir(), "package.json")
		invalidContent := `{invalid-json}`
		err := os.WriteFile(invalidManifestPath, []byte(invalidContent), 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		targetSpec := manifest.PackageSpec{
			Name:    "pnpm",
			Version: "9.2.0",
			Hash:    "+sha512.123456",
		}

		_, err = manifest.WritePackageSpec(invalidManifestPath, targetSpec)
		if !errors.Is(err, manifest.ErrInvalidManifest) {
			t.Fatalf("expected ErrInvalidManifest, got: %v", err)
		}
	})
}
