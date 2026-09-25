// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

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
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-isatty"

	"github.com/jadmadi/thermal/internal/thermal"
	"github.com/jadmadi/thermal/internal/version"
)

const (
	githubAPI      = "https://api.github.com/repos/jadmadi/thermal/releases/latest"
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
		_ = saveUpdateCache(updateCache{CheckedAt: time.Now().UTC(), LatestVersion: release.TagName})
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

	printUpgradeSuccess(os.Stdout, current, release.TagName)
	_ = saveUpdateCache(updateCache{CheckedAt: time.Now().UTC(), LatestVersion: release.TagName})
	return 0
}

// isBreakingUpgrade reports whether upgrading from current to latest crosses a major
// boundary (in 1.x+) or a minor boundary (in 0.x.y), indicating potential breaking changes.
func isBreakingUpgrade(current, latest string) bool {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")
	if current == "dev" || current == "" {
		return true
	}
	if current == latest {
		return false
	}
	curParts := strings.Split(strings.Split(current, "-")[0], ".")
	latParts := strings.Split(strings.Split(latest, "-")[0], ".")
	if len(curParts) < 2 || len(latParts) < 2 {
		return true
	}
	curMajor, err1 := strconv.Atoi(curParts[0])
	curMinor, err2 := strconv.Atoi(curParts[1])
	latMajor, err3 := strconv.Atoi(latParts[0])
	latMinor, err4 := strconv.Atoi(latParts[1])
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return true
	}
	if curMajor == 0 || latMajor == 0 {
		return curMajor != latMajor || curMinor != latMinor
	}
	return curMajor != latMajor
}

// printUpgradeSuccess formats and outputs upgrade completion and conditional migration notice.
func printUpgradeSuccess(w io.Writer, current, latestTag string) {
	fmt.Fprintf(w, "  ✓ upgraded to %s\n", latestTag)
	fmt.Fprintf(w, "  restart thermal to use the new version.\n")
	latestVersion := strings.TrimPrefix(latestTag, "v")
	if isBreakingUpgrade(current, latestVersion) {
		fmt.Fprintf(w, "  migration guide: https://jadmadi.net/projects/thermal/migration or docs/MIGRATION.md\n")
	}
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

	if _, err = io.Copy(out, resp.Body); err != nil {
		return err
	}
	return out.Close()
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

		// Look for the binary named "thermal".
		if filepath.Base(hdr.Name) == "thermal" {
			out, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer out.Close()
			if _, err = io.Copy(out, io.LimitReader(tr, maxBinarySize)); err != nil {
				return err
			}
			return out.Close()
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

type updateCache struct {
	CheckedAt     time.Time `json:"checked_at"`
	LatestVersion string    `json:"latest_version"`
}

func updateCachePath() string {
	home := thermal.HomeDir()
	return filepath.Join(home, ".cache", "thermal", "update.json")
}

func loadUpdateCache() (updateCache, error) {
	p := updateCachePath()
	b, err := os.ReadFile(p)
	if err != nil {
		return updateCache{}, err
	}
	var c updateCache
	if err := json.Unmarshal(b, &c); err != nil {
		return updateCache{}, err
	}
	return c, nil
}

func saveUpdateCache(c updateCache) error {
	p := updateCachePath()
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o644)
}

// semverCompare compares two semver strings (vX.Y.Z or X.Y.Z).
// Returns 1 if v1 > v2, -1 if v1 < v2, 0 if equal.
func semverCompare(v1, v2 string) int {
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")
	if v1 == v2 {
		return 0
	}
	if v1 == "dev" || v1 == "" {
		return -1
	}
	if v2 == "dev" || v2 == "" {
		return 1
	}
	p1 := strings.Split(strings.Split(v1, "-")[0], ".")
	p2 := strings.Split(strings.Split(v2, "-")[0], ".")
	for i := 0; i < 3; i++ {
		var n1, n2 int
		if i < len(p1) {
			n1, _ = strconv.Atoi(p1[i])
		}
		if i < len(p2) {
			n2, _ = strconv.Atoi(p2[i])
		}
		if n1 > n2 {
			return 1
		}
		if n1 < n2 {
			return -1
		}
	}
	return 0
}

func fetchLatestReleaseTag() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubAPI, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "thermal-update-check")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %s", resp.Status)
	}

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}
	return rel.TagName, nil
}

func runBackgroundUpdateCheck() {
	tag, err := fetchLatestReleaseTag()
	if err != nil || tag == "" {
		return
	}
	_ = saveUpdateCache(updateCache{
		CheckedAt:     time.Now().UTC(),
		LatestVersion: tag,
	})
}

func spawnBackgroundUpdateCheck() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	if strings.HasSuffix(exe, ".test") || strings.Contains(exe, "__debug_bin") {
		return
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		exe, _ = os.Executable()
	}

	cmd := exec.Command(exe, "--check-update-bg")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	setDetachedProcess(cmd)
	_ = cmd.Start()
}

func maybeCheckForUpdate(opts thermal.Options) {
	if opts.JSON || opts.Offline || opts.NoUpdateCheck {
		return
	}
	if os.Getenv("THERMAL_NO_UPDATE_CHECK") != "" || os.Getenv("CI") != "" {
		return
	}
	if opts.Tool == "upgrade" || opts.Tool == "--check-update-bg" {
		return
	}
	// Do not nag if not in an interactive terminal
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsTerminal(os.Stderr.Fd()) {
		return
	}

	cache, err := loadUpdateCache()
	now := time.Now().UTC()

	// If cache has a newer version, notify user
	current := version.Version
	if current != "dev" && current != "" && cache.LatestVersion != "" {
		if semverCompare(cache.LatestVersion, current) > 0 {
			curTag := "v" + strings.TrimPrefix(current, "v")
			latTag := "v" + strings.TrimPrefix(cache.LatestVersion, "v")
			msg := fmt.Sprintf("A new version of thermal is available: %s → %s. Run 'thermal upgrade' to update.", curTag, latTag)
			if !opts.NoColor {
				fmt.Fprintf(os.Stderr, "\n\033[2m%s\033[0m\n", msg)
			} else {
				fmt.Fprintf(os.Stderr, "\n%s\n", msg)
			}
		}
	}

	// If never checked, or last check is older than 24 hours, trigger background check
	if err != nil || now.Sub(cache.CheckedAt) > 24*time.Hour {
		// Update checked_at timestamp to avoid stampedes
		_ = saveUpdateCache(updateCache{
			CheckedAt:     now,
			LatestVersion: cache.LatestVersion,
		})
		spawnBackgroundUpdateCheck()
	}
}
