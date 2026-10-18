package version

import "runtime/debug"

func SetReadBuildInfoForTest(stub func() (*debug.BuildInfo, bool)) func() {
	originalReadBuildInfo := readBuildInfo
	readBuildInfo = stub
	return func() {
		readBuildInfo = originalReadBuildInfo
	}
}
