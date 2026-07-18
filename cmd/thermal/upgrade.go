package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jadmadi/thermal/internal/version"
)

const (
	githubAPI    = "https://api.github.com/repos/jadmadi/thermal/releases/latest"
	githubReleases = "https://github.com/jadmadi/thermal/releases"
)

type githubRelease struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	HTMLURL string  `json:"html_url"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

var httpClient = &http.Client{Timeout: 30 * time.Second}

// runUpgrade checks for a newer release and self-replaces the binary.
func runUpgrade() int {
	fmt.Printf("  thermal %s  checking for updates...\n", version.String())

	current := strings.TrimPrefix(version.Version, "v")

	// Fetch latest release from GitHub API.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubAPI, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: cannot create request: %v\n", err)
		return 1
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: cannot reach GitHub API: %v\n", err)
		return 1
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Fprintf(os.Stderr, "  error: GitHub API returned %s\n", resp.Status)
		return 1
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fmt.Fprintf(os.Stderr, "  error: cannot parse release info: %v\n", err)
		return 1
	}

	latest := strings.TrimPrefix(release.TagName, "v")

	if current == "dev" {
		fmt.Printf("  current: dev (built from source)\n")
		fmt.Printf("  latest:  %s\n", release.TagName)
	} else if current == latest {
		fmt.Printf("  already up to date — %s\n", release.TagName)
		return 0
	} else {
		fmt.Printf("  update available: %s → %s\n", "v"+current, release.TagName)
	}

	// Find the matching asset for this OS/arch.
	assetName, downloadURL, err := findAsset(release.Assets)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		fmt.Fprintf(os.Stderr, "  download manually: %s\n", release.HTMLURL)
		return 1
	}

	fmt.Printf("  downloading %s...\n", assetName)

	// Download the archive.
	tmpDir, err := os.MkdirTemp("", "thermal-upgrade-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: cannot create temp dir: %v\n", err)
		return 1
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, assetName)
	if err := downloadFile(downloadURL, archivePath); err != nil {
		fmt.Fprintf(os.Stderr, "  error: download failed: %v\n", err)
		return 1
	}

	// Extract the binary from the archive.
	binaryPath := filepath.Join(tmpDir, "thermal")
	if err := extractBinary(archivePath, assetName, binaryPath); err != nil {
		fmt.Fprintf(os.Stderr, "  error: extraction failed: %v\n", err)
		return 1
	}

	// Make it executable (in case extraction lost the mode).
	if err := os.Chmod(binaryPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "  error: cannot chmod: %v\n", err)
		return 1
	}

	// Find the current binary path.
	currentBin, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: cannot find current binary: %v\n", err)
		return 1
	}

	// Resolve symlinks.
	currentBin, err = filepath.EvalSymlinks(currentBin)
	if err != nil {
		currentBin, _ = os.Executable()
	}

	// Atomic swap: write to a temp file next to the target, then rename.
	oldPath := currentBin + ".old"
	tmpPath := currentBin + ".new"

	// Copy the new binary to tmpPath.
	if err := copyFile(binaryPath, tmpPath); err != nil {
		fmt.Fprintf(os.Stderr, "  error: cannot write new binary: %v\n", err)
		return 1
	}
	if err := os.Chmod(tmpPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "  error: cannot chmod: %v\n", err)
		return 1
	}

	// Rename current → old, new → current.
	_ = os.Remove(oldPath)
	os.Rename(currentBin, oldPath)
	if err := os.Rename(tmpPath, currentBin); err != nil {
		// Try to restore.
		os.Rename(oldPath, currentBin)
		fmt.Fprintf(os.Stderr, "  error: cannot replace binary: %v\n", err)
		return 1
	}

	// Clean up the old binary (best-effort — may fail on Windows if locked).
	_ = os.Remove(oldPath)

	fmt.Printf("  ✓ upgraded to %s\n", release.TagName)
	fmt.Printf("  restart thermal to use the new version.\n")
	return 0
}

// findAsset finds the release asset matching the current OS and architecture.
// Returns the asset name and its download URL.
func findAsset(assets []asset) (string, string, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Map Go arch names to goreleaser arch names (they match, but be safe).
	archMap := map[string]string{
		"amd64": "amd64",
		"arm64": "arm64",
		"386":   "386",
	}
	archName, ok := archMap[goarch]
	if !ok {
		archName = goarch
	}

	// goreleaser naming: thermal_VERSION_OS_ARCH.tar.gz (or .zip for windows)
	for _, a := range assets {
		name := strings.ToLower(a.Name)
		osMatch := strings.Contains(name, "_"+goos+"_") || strings.Contains(name, "-"+goos+"-")
		archMatch := strings.Contains(name, "_"+archName+".") || strings.Contains(name, "_"+archName+"_")

		if osMatch && archMatch {
			return a.Name, a.BrowserDownloadURL, nil
		}
	}

	return "", "", fmt.Errorf("no binary found for %s/%s", goos, goarch)
}

// downloadFile downloads a URL to a local path.
func downloadFile(url, dest string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractBinary extracts the thermal binary from a .tar.gz or .zip archive.
func extractBinary(archivePath, archiveName, destPath string) error {
	if strings.HasSuffix(archiveName, ".zip") {
		return fmt.Errorf("zip extraction not yet supported — please extract manually")
	}

	// .tar.gz
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	const maxBinarySize = 250 * 1024 * 1024 // 250 MB
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Look for the binary named "thermal" (not in a subdirectory).
		if filepath.Base(hdr.Name) == "thermal" && !strings.Contains(hdr.Name, "/") {
			out, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer out.Close()
			_, err = io.Copy(out, io.LimitReader(tr, maxBinarySize))
			return err
		}

		// Also handle if it's in a subdirectory.
		if filepath.Base(hdr.Name) == "thermal" {
			out, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer out.Close()
			_, err = io.Copy(out, io.LimitReader(tr, maxBinarySize))
			return err
		}
	}

	return fmt.Errorf("binary 'thermal' not found in archive")
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
