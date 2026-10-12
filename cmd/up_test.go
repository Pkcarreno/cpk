package cmd_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkcarreno/cpk/cmd"
	"github.com/pkcarreno/cpk/internal/manifest"
)

func TestNewUpCommand(t *testing.T) {
	t.Run("fails when the package.json file has no packageManager field", func(t *testing.T) {
		temporalDirectory := t.TempDir()
		manifestPath := filepath.Join(temporalDirectory, "package.json")
		err := os.WriteFile(manifestPath, []byte(`{"name": "test-pkg"}`), 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		t.Chdir(temporalDirectory)

		var output bytes.Buffer

		upCommand := cmd.NewUpCommand()
		upCommand.SetOut(&output)
		upCommand.SetErr(&output)

		err = upCommand.Execute()
		if !errors.Is(err, manifest.ErrMissingPackageManager) {
			t.Fatalf("expected ErrMissingPackageManager, got: %v", err)
		}
	})

	testCases := []struct {
		name                  string
		initialSpec           string
		arguments             []string
		expectedPackageName   string
		expectedVersion       string
		expectedIntegrityHash string
	}{
		{
			name:                  "updates to the latest stable when run with no arguments",
			initialSpec:           "pnpm@9.1.0+sha512.01010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101",
			arguments:             []string{},
			expectedPackageName:   "pnpm",
			expectedVersion:       "9.5.0",
			expectedIntegrityHash: "+sha512.8456c09bcf382e94fc13523eb08da1715a237027147feca54f962bdf8537a917c7fffda702dec6d4d1589aa1930e944f9cbd7166733122b03dff1b9ab5bb995d",
		},
		{
			name:                  "updates to the highest version in major prefix",
			initialSpec:           "pnpm@9.1.0+sha512.01010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101",
			arguments:             []string{"9"},
			expectedPackageName:   "pnpm",
			expectedVersion:       "9.5.0",
			expectedIntegrityHash: "+sha512.8456c09bcf382e94fc13523eb08da1715a237027147feca54f962bdf8537a917c7fffda702dec6d4d1589aa1930e944f9cbd7166733122b03dff1b9ab5bb995d",
		},
		{
			name:                  "updates to the highest version in minor prefix",
			initialSpec:           "pnpm@9.1.0+sha512.01010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101",
			arguments:             []string{"9.2"},
			expectedPackageName:   "pnpm",
			expectedVersion:       "9.2.4",
			expectedIntegrityHash: "+sha512.8772e32fdcf58c82b5210d721d59aa62c87adbef677da9f0327dc7f4c65b6c729a34f196f839722ed7a959de3868f399c77f568671a4a14edf8e80db5dc792f4",
		},
		{
			name:                  "updates to the exact version when full version is specified",
			initialSpec:           "pnpm@9.1.0+sha512.01010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101",
			arguments:             []string{"9.2.0"},
			expectedPackageName:   "pnpm",
			expectedVersion:       "9.2.0",
			expectedIntegrityHash: "+sha512.1976f888a31564e233e924c43c596a24c010726cce59fa8085dfb67fb5253cbe2acdaf7ff7b557e3cdca03317aca14b7ad188f00622cec6aec1b5fec4233a91a",
		},
		{
			name:                  "updates to the version published under dist-tag",
			initialSpec:           "pnpm@9.1.0+sha512.01010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101",
			arguments:             []string{"next"},
			expectedPackageName:   "pnpm",
			expectedVersion:       "10.0.0-beta.0",
			expectedIntegrityHash: "+sha512.2a798e827929dd930fc9db22b85bc33895be31e5038c80954ad9a9a09a6a2c4633e7fd07e49070d793a97723da9165f6f27016b99f1f9fde801aaab5f8e77f1a",
		},
		{
			name:                  "upgrades yarn classic to highest 1.x release when run with no arguments",
			initialSpec:           "yarn@1.22.19+sha1.0101010101010101010101010101010101010101",
			arguments:             []string{},
			expectedPackageName:   "yarn",
			expectedVersion:       "1.22.22",
			expectedIntegrityHash: "+sha1.ac34549e6aa8e7ead463a7407e1c7390f61a6610",
		},
		{
			name:                  "upgrades yarn berry to latest stable release when run with no arguments",
			initialSpec:           "yarn@4.1.0+sha512.01010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101",
			arguments:             []string{},
			expectedPackageName:   "yarn",
			expectedVersion:       "4.18.0",
			expectedIntegrityHash: "+sha512.8456c09bcf382e94fc13523eb08da1715a237027147feca54f962bdf8537a917c7fffda702dec6d4d1589aa1930e944f9cbd7166733122b03dff1b9ab5bb995d",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			temporalDirectory := t.TempDir()
			manifestPath := filepath.Join(temporalDirectory, "package.json")
			initialContent := fmt.Sprintf(`{"name": "test-pkg","packageManager": "%s"}`, testCase.initialSpec)
			err := os.WriteFile(manifestPath, []byte(initialContent), 0600)
			if err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			t.Chdir(temporalDirectory)

			mockServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				responseWriter.Header().Set("Content-Type", "application/json")
				switch request.URL.Path {
				case "/pnpm":
					responseWriter.WriteHeader(http.StatusOK)
					responseWriter.Write([]byte(`{
                        "name": "pnpm",
                        "dist-tags": {
                            "latest": "9.5.0",
                            "next": "10.0.0-beta.0"
                        },
                        "versions": {
                            "9.1.0": { "dist": { "integrity": "sha512-AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ==" } },
                            "9.2.0": { "dist": { "integrity": "sha512-GXb4iKMVZOIz6STEPFlqJMAQcmzOWfqAhd+2f7UlPL4qza9/97VX483KAzF6yhS3rRiPAGIs7GrsG1/sQjOpGg==" } },
                            "9.2.4": { "dist": { "integrity": "sha512-h3LjL9z1jIK1IQ1yHVmqYsh62+9nfanwMn3H9MZbbHKaNPGW+DlyLtepWd44aPOZx39WhnGkoU7fjoDbXceS9A==" } },
                            "9.5.0": { "dist": { "integrity": "sha512-hFbAm884LpT8E1I+sI2hcVojcCcUf+ylT5Yr34U3qRfH//2nAt7G1NFYmqGTDpRPnL1xZnMxIrA9/xuatbuZXQ==" } },
                            "10.0.0-beta.0": { "dist": { "integrity": "sha512-KnmOgnkp3ZMPydsiuFvDOJW+MeUDjICVStmpoJpqLEYz5/0H5JBw15OpdyPakWX28nAWuZ8fn96AGqq1+Od/Gg==" } }
                        }
                    }`))
				case "/@yarnpkg/cli-dist":
					responseWriter.WriteHeader(http.StatusOK)
					responseWriter.Write([]byte(`{
                        "name": "@yarnpkg/cli-dist",
                        "dist-tags": {"latest": "4.18.0"},
                        "versions": {
                            "4.1.0": {"dist": {"integrity": "sha512-AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ=="}},
                            "4.18.0": {"dist": {"integrity": "sha512-hFbAm884LpT8E1I+sI2hcVojcCcUf+ylT5Yr34U3qRfH//2nAt7G1NFYmqGTDpRPnL1xZnMxIrA9/xuatbuZXQ=="}}
                        }
                    }`))
				case "/yarn":
					responseWriter.WriteHeader(http.StatusOK)
					responseWriter.Write([]byte(`{
                        "name": "yarn",
                        "dist-tags": {"latest": "1.22.22"},
                        "versions": {
                            "1.22.19": {"dist": {"shasum": "0101010101010101010101010101010101010101"}},
                            "1.22.22": {"dist": {"shasum": "ac34549e6aa8e7ead463a7407e1c7390f61a6610"}},
                            "2.0.0": {"dist": {"shasum": "0202020202020202020202020202020202020202"}}
                        }
                    }`))
				default:
					responseWriter.WriteHeader(http.StatusNotFound)
				}
			}))
			defer mockServer.Close()

			t.Setenv("NPM_CONFIG_REGISTRY", mockServer.URL)

			var output bytes.Buffer

			upCommand := cmd.NewUpCommand()
			upCommand.SetArgs(testCase.arguments)
			upCommand.SetOut(&output)
			upCommand.SetErr(&output)

			err = upCommand.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			initialSpec, err := manifest.ParsePackageSpec(testCase.initialSpec)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			expectedOutput := fmt.Sprintf("Updated %s from %s to %s\n", initialSpec.Name, initialSpec.Version, testCase.expectedVersion)

			if output.String() != expectedOutput {
				t.Fatalf("output does not match expected output, expected: %q, got: %v", expectedOutput, output)
			}

			updatedPackageSpec, err := manifest.ReadPackageSpec(manifestPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			expectedPackageSpec := manifest.PackageSpec{
				Name:    testCase.expectedPackageName,
				Version: testCase.expectedVersion,
				Hash:    testCase.expectedIntegrityHash,
			}

			if updatedPackageSpec != expectedPackageSpec {
				t.Fatalf("updated package spec does not match expected package spec, expected: %q, got: %q", expectedPackageSpec.String(), updatedPackageSpec.String())
			}
		})
	}

	t.Run("preserves the manifest file when the version and integrity hash already match", func(t *testing.T) {
		temporalDirectory := t.TempDir()
		manifestPath := filepath.Join(temporalDirectory, "package.json")
		initialContent := []byte(`{"name": "test-pkg","packageManager": "pnpm@9.5.0+sha512.8456c09bcf382e94fc13523eb08da1715a237027147feca54f962bdf8537a917c7fffda702dec6d4d1589aa1930e944f9cbd7166733122b03dff1b9ab5bb995d"}`)
		err := os.WriteFile(manifestPath, initialContent, 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		t.Chdir(temporalDirectory)

		mockServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.Header().Set("Content-Type", "application/json")
			responseWriter.WriteHeader(http.StatusOK)
			responseWriter.Write([]byte(`{
					"name": "pnpm",
					"dist-tags": {"latest": "9.5.0"},
					"versions": {
						"9.1.0": {"dist": {"integrity": "sha512-AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ=="}},
						"9.5.0": {"dist": {"integrity": "sha512-hFbAm884LpT8E1I+sI2hcVojcCcUf+ylT5Yr34U3qRfH//2nAt7G1NFYmqGTDpRPnL1xZnMxIrA9/xuatbuZXQ=="}}
					}
				}`))
		}))
		defer mockServer.Close()

		t.Setenv("NPM_CONFIG_REGISTRY", mockServer.URL)

		var output bytes.Buffer

		upCommand := cmd.NewUpCommand()
		upCommand.SetArgs([]string{"9.5.0"})
		upCommand.SetOut(&output)
		upCommand.SetErr(&output)

		beforeFileInfo, err := os.Stat(manifestPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = upCommand.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		afterFileInfo, err := os.Stat(manifestPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !afterFileInfo.ModTime().Equal(beforeFileInfo.ModTime()) {
			t.Fatalf("expected the file modification time to remain unchanged")
		}

		expectedOutput := "pnpm is already up to date (9.5.0)\n"

		if output.String() != expectedOutput {
			t.Fatalf("output does not match expected output, expected: %q, got: %v", expectedOutput, output)
		}

		updatedPackageSpec, err := manifest.ReadPackageSpec(manifestPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedPackageSpec := manifest.PackageSpec{
			Name:    "pnpm",
			Version: "9.5.0",
			Hash:    "+sha512.8456c09bcf382e94fc13523eb08da1715a237027147feca54f962bdf8537a917c7fffda702dec6d4d1589aa1930e944f9cbd7166733122b03dff1b9ab5bb995d",
		}

		if updatedPackageSpec != expectedPackageSpec {
			t.Fatalf("updated package spec does not match expected package spec, expected: %q, got: %q", expectedPackageSpec.String(), updatedPackageSpec.String())
		}
	})

	t.Run("returns an error when the requested version is lower than the current one", func(t *testing.T) {
		temporalDirectory := t.TempDir()
		manifestPath := filepath.Join(temporalDirectory, "package.json")
		initialContent := []byte(`{"name": "test-pkg","packageManager": "pnpm@9.5.0+sha512.8456c09bcf382e94fc13523eb08da1715a237027147feca54f962bdf8537a917c7fffda702dec6d4d1589aa1930e944f9cbd7166733122b03dff1b9ab5bb995d"}`)
		err := os.WriteFile(manifestPath, initialContent, 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		t.Chdir(temporalDirectory)

		mockServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.Header().Set("Content-Type", "application/json")
			responseWriter.WriteHeader(http.StatusOK)
			responseWriter.Write([]byte(`{
					"name": "pnpm",
					"dist-tags": {"latest": "9.5.0"},
					"versions": {
						"9.1.0": {"dist": {"integrity": "sha512-AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ=="}},
						"9.5.0": {"dist": {"integrity": "sha512-hFbAm884LpT8E1I+sI2hcVojcCcUf+ylT5Yr34U3qRfH//2nAt7G1NFYmqGTDpRPnL1xZnMxIrA9/xuatbuZXQ=="}}
					}
				}`))
		}))
		defer mockServer.Close()

		t.Setenv("NPM_CONFIG_REGISTRY", mockServer.URL)

		var output bytes.Buffer

		upCommand := cmd.NewUpCommand()
		upCommand.SetArgs([]string{"9.1.0"})
		upCommand.SetOut(&output)
		upCommand.SetErr(&output)

		err = upCommand.Execute()
		if !errors.Is(err, cmd.ErrDowngrade) {
			t.Fatalf("expected ErrDowngrade, got: %v", err)
		}
	})

	t.Run("preserves the manifest file when the pinned yarn classic version already matches", func(t *testing.T) {
		temporalDirectory := t.TempDir()
		manifestPath := filepath.Join(temporalDirectory, "package.json")
		initialContent := []byte(`{"name": "test-pkg","packageManager": "yarn@1.22.22+sha1.ac34549e6aa8e7ead463a7407e1c7390f61a6610"}`)
		err := os.WriteFile(manifestPath, initialContent, 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		t.Chdir(temporalDirectory)

		mockServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.Header().Set("Content-Type", "application/json")
			switch request.URL.Path {
			case "/yarn":
				responseWriter.WriteHeader(http.StatusOK)
				responseWriter.Write([]byte(`{
                    "name": "yarn",
                    "dist-tags": {"latest": "1.22.22"},
                    "versions": {
                        "1.22.22": {"dist": {"shasum": "ac34549e6aa8e7ead463a7407e1c7390f61a6610"}}
                    }
                }`))
			default:
				responseWriter.WriteHeader(http.StatusNotFound)
			}
		}))
		defer mockServer.Close()

		t.Setenv("NPM_CONFIG_REGISTRY", mockServer.URL)

		var output bytes.Buffer

		upCommand := cmd.NewUpCommand()
		upCommand.SetArgs([]string{})
		upCommand.SetOut(&output)
		upCommand.SetErr(&output)

		beforeFileInfo, err := os.Stat(manifestPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = upCommand.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		afterFileInfo, err := os.Stat(manifestPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !afterFileInfo.ModTime().Equal(beforeFileInfo.ModTime()) {
			t.Fatalf("expected the file modification time to remain unchanged")
		}

		expectedOutput := "yarn is already up to date (1.22.22)\n"
		if output.String() != expectedOutput {
			t.Fatalf("output does not match expected output, expected: %q, got: %v", expectedOutput, output)
		}
	})
}
