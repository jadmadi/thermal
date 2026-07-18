package main

import (
	"runtime"
	"testing"
)

func TestFindAsset_ExactMatching(t *testing.T) {
	assets := []asset{
		{Name: "thermal_0.3.0_linux_amd64.tar.gz", BrowserDownloadURL: "https://example.com/linux_amd64"},
		{Name: "thermal_0.3.0_darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/darwin_arm64"},
	}

	name, _, err := findAsset(assets)
	if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		if err != nil || name != "thermal_0.3.0_linux_amd64.tar.gz" {
			t.Errorf("failed matching linux_amd64: %v", err)
		}
	} else if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		if err != nil || name != "thermal_0.3.0_darwin_arm64.tar.gz" {
			t.Errorf("failed matching darwin_arm64: %v", err)
		}
	} else {
		// If on windows or other arch, must return clean error without picking wrong architecture
		if err == nil {
			t.Errorf("expected error when no matching architecture exists, got %q", name)
		}
	}
}
