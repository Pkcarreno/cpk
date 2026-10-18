package version_test

import (
	"runtime"
	"runtime/debug"
	"testing"

	"github.com/pkcarreno/cpk/internal/version"
)

func TestCurrent(t *testing.T) {
	testCases := []struct {
		name        string
		buildInfo   *debug.BuildInfo
		hasInfo     bool
		wantVersion string
		wantCommit  string
		wantDate    string
	}{
		{
			name: "extracts VCS metadata without dirty suffix when working tree is clean",
			buildInfo: &debug.BuildInfo{
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "f4f7f00b8d48fe6d8f40df2fe82d7c13e782f9fd"},
					{Key: "vcs.time", Value: "2026-10-17T18:21:47Z"},
					{Key: "vcs.modified", Value: "false"},
				},
			},
			hasInfo:     true,
			wantVersion: "dev",
			wantCommit:  "f4f7f00b8d48fe6d8f40df2fe82d7c13e782f9fd",
			wantDate:    "2026-10-17T18:21:47Z",
		},
		{
			name: "appends dirty suffix when repository has modified files",
			buildInfo: &debug.BuildInfo{
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "f4f7f00b8d48fe6d8f40df2fe82d7c13e782f9fd"},
					{Key: "vcs.time", Value: "2026-10-17T18:21:47Z"},
					{Key: "vcs.modified", Value: "true"},
				},
			},
			hasInfo:     true,
			wantVersion: "dev",
			wantCommit:  "f4f7f00b8d48fe6d8f40df2fe82d7c13e782f9fd-dirty",
			wantDate:    "2026-10-17T18:21:47Z",
		},
		{
			name: "uses main module version when populated",
			buildInfo: &debug.BuildInfo{
				Main: debug.Module{Version: "v0.2.0"},
			},
			hasInfo:     true,
			wantVersion: "0.2.0",
		},
		{
			name: "ignores devel main module version",
			buildInfo: &debug.BuildInfo{
				Main: debug.Module{Version: "(devel)"},
			},
			hasInfo:     true,
			wantVersion: "dev",
		},
		{
			name:        "falls back to dev version when build info is unavailable",
			buildInfo:   nil,
			hasInfo:     false,
			wantVersion: "dev",
			wantCommit:  "",
			wantDate:    "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			restore := version.SetReadBuildInfoForTest(func() (*debug.BuildInfo, bool) {
				return testCase.buildInfo, testCase.hasInfo
			})
			defer restore()

			buildMetadata := version.Current()

			if buildMetadata.Version != testCase.wantVersion {
				t.Fatalf("version = %q, want %q", buildMetadata.Version, testCase.wantVersion)
			}
			if buildMetadata.Commit != testCase.wantCommit {
				t.Fatalf("commit = %q, want %q", buildMetadata.Commit, testCase.wantCommit)
			}
			if buildMetadata.Date != testCase.wantDate {
				t.Fatalf("date = %q, want %q", buildMetadata.Date, testCase.wantDate)
			}
			if buildMetadata.Go != runtime.Version() {
				t.Fatalf("go = %q, want %q", buildMetadata.Go, runtime.Version())
			}
			if buildMetadata.OS != runtime.GOOS {
				t.Fatalf("os = %q, want %q", buildMetadata.OS, runtime.GOOS)
			}
			if buildMetadata.Arch != runtime.GOARCH {
				t.Fatalf("arch = %q, want %q", buildMetadata.Arch, runtime.GOARCH)
			}
			if buildMetadata.Compiler != runtime.Compiler {
				t.Fatalf("compiler = %q, want %q", buildMetadata.Compiler, runtime.Compiler)
			}
		})
	}
}

func TestBuildMetadata_String(t *testing.T) {
	t.Run("renders multiline diagnostic metadata with all fields", func(t *testing.T) {
		buildMetadata := version.BuildMetadata{
			Version:  "0.1.0",
			Commit:   "f4f7f00b8d48fe6d8f40df2fe82d7c13e782f9fd",
			Date:     "2026-10-17T18:21:47Z",
			Go:       "go1.26.5",
			OS:       "linux",
			Arch:     "amd64",
			Compiler: "gc",
		}

		want := "cpk 0.1.0\ncommit f4f7f00b8d48fe6d8f40df2fe82d7c13e782f9fd\ndate 2026-10-17T18:21:47Z\ngo go1.26.5 linux/amd64\n"

		if got := buildMetadata.String(); got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("omits commit and date lines when fields are empty", func(t *testing.T) {
		buildMetadata := version.BuildMetadata{
			Version:  "0.1.0",
			Commit:   "",
			Date:     "",
			Go:       "go1.26.5",
			OS:       "linux",
			Arch:     "amd64",
			Compiler: "gc",
		}

		want := "cpk 0.1.0\ngo go1.26.5 linux/amd64\n"

		if got := buildMetadata.String(); got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})
}
