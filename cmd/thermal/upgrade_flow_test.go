// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func upgradeArchive(t *testing.T, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	body := []byte(content)
	if err := tw.WriteHeader(&tar.Header{Name: "thermal", Mode: 0o755, Size: int64(len(body))}); err != nil {
		t.Fatalf("tar header: %v", err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatalf("tar write: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar close: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	return buf.Bytes()
}

func upgradeServer(t *testing.T, tag string, archive []byte, downloadStatus int) (*httptest.Server, *int, *int) {
	t.Helper()
	var downloads, apiHits int
	mux := http.NewServeMux()
	mux.HandleFunc("/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		apiHits++
		if tag == "" {
			http.NotFound(w, r)
			return
		}
		assetName := fmt.Sprintf("thermal_%s_%s_%s.tar.gz", strings.TrimPrefix(tag, "v"), runtime.GOOS, runtime.GOARCH)
		base := "http://" + r.Host
		_ = json.NewEncoder(w).Encode(githubRelease{
			TagName: tag,
			HTMLURL: base + "/release",
			Assets:  []asset{{Name: assetName, BrowserDownloadURL: base + "/download"}},
		})
	})
	mux.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		downloads++
		if downloadStatus != http.StatusOK {
			w.WriteHeader(downloadStatus)
			return
		}
		_, _ = w.Write(archive)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, &downloads, &apiHits
}

func upgradeTarget(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "thermal")
	if err := os.WriteFile(p, []byte(content), 0o755); err != nil {
		t.Fatalf("write target: %v", err)
	}
	return p
}

func newUpgradeOptions(srv *httptest.Server, target, current string) upgradeOptions {
	return upgradeOptions{
		Out:            &bytes.Buffer{},
		Err:            &bytes.Buffer{},
		Stdin:          strings.NewReader(""),
		Client:         srv.Client(),
		DownloadClient: srv.Client(),
		APIURL:         srv.URL + "/releases/latest",
		Executable:     func() (string, error) { return target, nil },
		CurrentVersion: current,
		CurrentLabel:   current,
		TTY:            false,
		TTYSet:         true,
	}
}

func outputOf(w any) string {
	if b, ok := w.(*bytes.Buffer); ok {
		return b.String()
	}
	return ""
}

func TestUpgradeCheckDoesNotMutate(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	target := upgradeTarget(t, "old")
	srv, downloads, apiHits := upgradeServer(t, "v0.22.0", upgradeArchive(t, "new"), http.StatusOK)
	o := newUpgradeOptions(srv, target, "0.21.0")
	o.Check = true
	if code := runUpgradeWithOptions(o); code != 0 {
		t.Fatalf("check exit=%d err=%s", code, outputOf(o.Err))
	}
	if *downloads != 0 {
		t.Fatalf("check downloaded %d times", *downloads)
	}
	if *apiHits != 1 {
		t.Fatalf("api hits=%d", *apiHits)
	}
	if got := readFile(t, target); got != "old" {
		t.Fatalf("target mutated: %q", got)
	}
	out := outputOf(o.Out)
	if !strings.Contains(out, "action: check") || !strings.Contains(out, "thermal_0.22.0") {
		t.Fatalf("unexpected check output:\n%s", out)
	}
}

func TestUpgradeDryRunDoesNotMutate(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	target := upgradeTarget(t, "old")
	srv, downloads, _ := upgradeServer(t, "v0.22.0", upgradeArchive(t, "new"), http.StatusOK)
	o := newUpgradeOptions(srv, target, "0.21.0")
	o.DryRun = true
	if code := runUpgradeWithOptions(o); code != 0 {
		t.Fatalf("dry-run exit=%d err=%s", code, outputOf(o.Err))
	}
	if *downloads != 0 {
		t.Fatalf("dry-run downloaded %d times", *downloads)
	}
	if got := readFile(t, target); got != "old" {
		t.Fatalf("target mutated: %q", got)
	}
	if !strings.Contains(outputOf(o.Out), "action: dry-run") {
		t.Fatalf("unexpected dry-run output:\n%s", outputOf(o.Out))
	}
}

func TestUpgradeJSONContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	target := upgradeTarget(t, "old")
	srv, _, _ := upgradeServer(t, "v0.22.0", upgradeArchive(t, "new"), http.StatusOK)
	o := newUpgradeOptions(srv, target, "0.21.0")
	o.Check = true
	o.JSON = true
	if code := runUpgradeWithOptions(o); code != 0 {
		t.Fatalf("json check exit=%d err=%s", code, outputOf(o.Err))
	}
	var payload upgradeResult
	if err := json.Unmarshal([]byte(outputOf(o.Out)), &payload); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, outputOf(o.Out))
	}
	if payload.Type != "upgrade" || payload.Action != "check" || payload.Mutated {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if payload.Asset == "" || payload.Target == "" || payload.Latest != "v0.22.0" {
		t.Fatalf("incomplete payload: %+v", payload)
	}
}

func TestUpgradeNonTTYRefusal(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	target := upgradeTarget(t, "old")
	srv, downloads, apiHits := upgradeServer(t, "v0.22.0", upgradeArchive(t, "new"), http.StatusOK)
	o := newUpgradeOptions(srv, target, "0.21.0")
	o.Yes = false
	if code := runUpgradeWithOptions(o); code != 3 {
		t.Fatalf("exit=%d, want 3", code)
	}
	if *downloads != 0 || *apiHits != 0 {
		t.Fatalf("refusal touched network: downloads=%d api=%d", *downloads, *apiHits)
	}
	if !strings.Contains(outputOf(o.Err), "refusing to replace the binary") {
		t.Fatalf("missing refusal: %s", outputOf(o.Err))
	}
	if got := readFile(t, target); got != "old" {
		t.Fatalf("target mutated: %q", got)
	}
}

func TestUpgradeJSONDeclined(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	target := upgradeTarget(t, "old")
	srv, _, apiHits := upgradeServer(t, "v0.22.0", upgradeArchive(t, "new"), http.StatusOK)
	o := newUpgradeOptions(srv, target, "0.21.0")
	o.JSON = true
	o.Yes = false
	if code := runUpgradeWithOptions(o); code != 3 {
		t.Fatalf("exit=%d, want 3", code)
	}
	if *apiHits != 0 {
		t.Fatalf("json refusal touched network: %d", *apiHits)
	}
	var payload upgradeResult
	if err := json.Unmarshal([]byte(outputOf(o.Out)), &payload); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, outputOf(o.Out))
	}
	if payload.Action != "declined" || !strings.Contains(payload.Error, "confirmation required") {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestUpgradeDeclinePrompt(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	target := upgradeTarget(t, "old")
	srv, downloads, _ := upgradeServer(t, "v0.22.0", upgradeArchive(t, "new"), http.StatusOK)
	o := newUpgradeOptions(srv, target, "0.21.0")
	o.TTY = true
	o.Stdin = strings.NewReader("n\n")
	if code := runUpgradeWithOptions(o); code != 3 {
		t.Fatalf("exit=%d, want 3", code)
	}
	if *downloads != 0 {
		t.Fatalf("decline downloaded %d times", *downloads)
	}
	if got := readFile(t, target); got != "old" {
		t.Fatalf("target mutated: %q", got)
	}
	if !strings.Contains(outputOf(o.Err), "Upgrade thermal") {
		t.Fatalf("missing prompt: %s", outputOf(o.Err))
	}
}

func TestUpgradeYesReplacesBinary(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	target := upgradeTarget(t, "old")
	srv, downloads, _ := upgradeServer(t, "v0.22.0", upgradeArchive(t, "new-binary"), http.StatusOK)
	o := newUpgradeOptions(srv, target, "0.21.0")
	o.Yes = true
	if code := runUpgradeWithOptions(o); code != 0 {
		t.Fatalf("exit=%d err=%s", code, outputOf(o.Err))
	}
	if *downloads != 1 {
		t.Fatalf("downloads=%d", *downloads)
	}
	if got := readFile(t, target); got != "new-binary" {
		t.Fatalf("target=%q", got)
	}
	if _, err := os.Stat(target + ".old"); !os.IsNotExist(err) {
		t.Fatalf(".old remains")
	}
	if _, err := os.Stat(target + ".new"); !os.IsNotExist(err) {
		t.Fatalf(".new remains")
	}
	if !strings.Contains(outputOf(o.Out), "✓ upgraded to v0.22.0") {
		t.Fatalf("missing success: %s", outputOf(o.Out))
	}
}

func TestUpgradeSameVersionNoMutation(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	target := upgradeTarget(t, "old")
	srv, downloads, _ := upgradeServer(t, "v0.21.0", upgradeArchive(t, "new"), http.StatusOK)
	o := newUpgradeOptions(srv, target, "0.21.0")
	o.TTY = true
	if code := runUpgradeWithOptions(o); code != 0 {
		t.Fatalf("exit=%d err=%s", code, outputOf(o.Err))
	}
	if *downloads != 0 {
		t.Fatalf("downloads=%d", *downloads)
	}
	if !strings.Contains(outputOf(o.Out), "already up to date") {
		t.Fatalf("unexpected output: %s", outputOf(o.Out))
	}
	if got := readFile(t, target); got != "old" {
		t.Fatalf("target mutated: %q", got)
	}
}

func TestUpgradeForceSameVersion(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	target := upgradeTarget(t, "old")
	srv, downloads, _ := upgradeServer(t, "v0.21.0", upgradeArchive(t, "new"), http.StatusOK)
	o := newUpgradeOptions(srv, target, "0.21.0")
	o.Force = true
	o.Yes = true
	if code := runUpgradeWithOptions(o); code != 0 {
		t.Fatalf("exit=%d err=%s", code, outputOf(o.Err))
	}
	if *downloads != 1 {
		t.Fatalf("downloads=%d", *downloads)
	}
	if got := readFile(t, target); got != "new" {
		t.Fatalf("target=%q", got)
	}
}

func TestUpgradeReleaseAPIFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	target := upgradeTarget(t, "old")
	srv, _, _ := upgradeServer(t, "", nil, http.StatusOK)
	o := newUpgradeOptions(srv, target, "0.21.0")
	o.Check = true
	if code := runUpgradeWithOptions(o); code != 1 {
		t.Fatalf("exit=%d", code)
	}
	if !strings.Contains(outputOf(o.Err), "release API returned") {
		t.Fatalf("missing API error: %s", outputOf(o.Err))
	}
	if got := readFile(t, target); got != "old" {
		t.Fatalf("target mutated: %q", got)
	}
}

func TestUpgradeTruncatedArchiveLeavesTarget(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	target := upgradeTarget(t, "old")
	srv, _, _ := upgradeServer(t, "v0.22.0", []byte("not a gzip archive"), http.StatusOK)
	o := newUpgradeOptions(srv, target, "0.21.0")
	o.Yes = true
	if code := runUpgradeWithOptions(o); code != 1 {
		t.Fatalf("exit=%d", code)
	}
	if got := readFile(t, target); got != "old" {
		t.Fatalf("target mutated: %q", got)
	}
	if _, err := os.Stat(target + ".new"); !os.IsNotExist(err) {
		t.Fatalf(".new remains")
	}
	if _, err := os.Stat(target + ".old"); !os.IsNotExist(err) {
		t.Fatalf(".old remains")
	}
}

func TestUpgradeUnwritableTargetDirectory(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	missingDir := filepath.Join(t.TempDir(), "missing")
	target := filepath.Join(missingDir, "thermal")
	srv, _, _ := upgradeServer(t, "v0.22.0", upgradeArchive(t, "new"), http.StatusOK)
	o := newUpgradeOptions(srv, target, "0.21.0")
	o.Yes = true
	if code := runUpgradeWithOptions(o); code != 1 {
		t.Fatalf("exit=%d", code)
	}
	if !strings.Contains(outputOf(o.Err), "target directory is not writable") {
		t.Fatalf("missing writable error: %s", outputOf(o.Err))
	}
}

func TestParseUpgradeArgs(t *testing.T) {
	cases := []struct {
		args    []string
		wantErr bool
		check   func(upgradeOptions) bool
	}{
		{[]string{"upgrade"}, false, func(o upgradeOptions) bool { return !o.Check && !o.DryRun }},
		{[]string{"upgrade", "--check", "--json"}, false, func(o upgradeOptions) bool { return o.Check && o.JSON }},
		{[]string{"--check", "--dry-run", "upgrade"}, false, func(o upgradeOptions) bool { return o.Check && !o.DryRun }},
		{[]string{"upgrade", "--yes", "--force"}, false, func(o upgradeOptions) bool { return o.Yes && o.Force }},
		{[]string{"upgrade", "--quiet"}, true, nil},
		{[]string{"upgrade", "daily"}, true, nil},
		{[]string{"upgrade", "--check=maybe"}, true, nil},
	}
	for _, tc := range cases {
		o, err := parseUpgradeArgs(tc.args)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("%v: expected error", tc.args)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%v: %v", tc.args, err)
		}
		if tc.check != nil && !tc.check(o) {
			t.Fatalf("%v: unexpected options %+v", tc.args, o)
		}
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}
