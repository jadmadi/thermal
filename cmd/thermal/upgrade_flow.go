// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-isatty"

	"github.com/jadmadi/thermal/internal/version"
)

// upgradeOptions is the injectable state for the upgrade command.
type upgradeOptions struct {
	Check, DryRun, JSON, Yes, Force, Help bool

	Out, Err io.Writer
	Stdin    io.Reader

	Client         *http.Client
	DownloadClient *http.Client
	APIURL         string
	Executable     func() (string, error)

	CurrentVersion string
	CurrentLabel   string
	TTY            bool
	TTYSet         bool
}

type upgradePlan struct {
	Current, CurrentLabel string
	LatestTag, Latest     string
	AssetName, AssetURL   string
	Target                string
	ReleaseURL            string
	UpdateAvailable       bool
	Breaking              bool
}

type upgradeResult struct {
	Type      string `json:"type"`
	Action    string `json:"action"`
	Current   string `json:"current"`
	Latest    string `json:"latest"`
	Asset     string `json:"asset,omitempty"`
	Target    string `json:"target,omitempty"`
	Mutated   bool   `json:"mutated"`
	Migration string `json:"migration,omitempty"`
	Error     string `json:"error,omitempty"`
}

// upgradeRequested reports whether the first positional is the upgrade verb.
func upgradeRequested(args []string) bool {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			return i+1 < len(args) && strings.EqualFold(args[i+1], "upgrade")
		}
		if strings.HasPrefix(a, "-") && a != "-" {
			name := a
			if eq := strings.Index(a, "="); eq >= 0 {
				name = a[:eq]
			}
			key := strings.TrimLeft(name, "-")
			if _, ok := flagMetas[key]; ok && !strings.Contains(a, "=") {
				i++
			}
			continue
		}
		return strings.EqualFold(a, "upgrade")
	}
	return false
}

func parseBoolFlag(name, val string, hasValue bool) (bool, error) {
	if !hasValue {
		return true, nil
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return false, &usageError{
			Message: fmt.Sprintf("%s wants true or false, got %q", name, val),
			Fix:     "run thermal upgrade --help",
		}
	}
	return b, nil
}

// parseUpgradeArgs parses only the upgrade command's flags. Unknown flags and
// stray positionals are usage errors.
func parseUpgradeArgs(args []string) (upgradeOptions, error) {
	var o upgradeOptions
	seenUpgrade := false

	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "upgrade" && !seenUpgrade {
			seenUpgrade = true
			continue
		}
		if a == "--" {
			if i+1 < len(args) {
				return o, &usageError{Message: fmt.Sprintf("unexpected argument %q", args[i+1]), Fix: "run thermal upgrade --help"}
			}
			break
		}
		if a == "-h" || a == "--help" {
			o.Help = true
			continue
		}

		name, val := a, ""
		hasValue := false
		if eq := strings.Index(a, "="); eq >= 0 {
			name, val, hasValue = a[:eq], a[eq+1:], true
		}
		setBool := func(dst *bool) error {
			b, err := parseBoolFlag(name, val, hasValue)
			if err != nil {
				return err
			}
			*dst = b
			return nil
		}
		switch name {
		case "--check":
			if err := setBool(&o.Check); err != nil {
				return o, err
			}
		case "--dry-run":
			if err := setBool(&o.DryRun); err != nil {
				return o, err
			}
		case "--json":
			if err := setBool(&o.JSON); err != nil {
				return o, err
			}
		case "--yes":
			if err := setBool(&o.Yes); err != nil {
				return o, err
			}
		case "--force":
			if err := setBool(&o.Force); err != nil {
				return o, err
			}
		case "--no-color", "-v", "--verbose", "-d":
			// Accepted global flags with no effect on upgrade.
		default:
			if strings.HasPrefix(a, "-") {
				return o, &usageError{
					Message: fmt.Sprintf("unknown flag for upgrade: %s", name),
					Fix:     "run thermal upgrade --help",
				}
			}
			return o, &usageError{
				Message: fmt.Sprintf("unexpected argument %q", a),
				Fix:     "run thermal upgrade --help",
			}
		}
	}
	if o.Check {
		o.DryRun = false
	}
	return o, nil
}

// runUpgrade parses args, fills production defaults, and runs the state machine.
func runUpgrade(args []string, out, errW io.Writer) int {
	parsed, err := parseUpgradeArgs(args)
	if err != nil {
		var ue *usageError
		if errors.As(err, &ue) {
			printUsageError(errW, ue)
		} else {
			fmt.Fprintf(errW, "thermal: %v\n", err)
		}
		return 2
	}
	if parsed.Help {
		return printHelpFor("upgrade", "", "", false)
	}

	parsed.Out = out
	parsed.Err = errW
	if parsed.Stdin == nil {
		parsed.Stdin = os.Stdin
	}
	if parsed.Client == nil {
		parsed.Client = httpClient
	}
	if parsed.DownloadClient == nil {
		parsed.DownloadClient = downloadClient
	}
	if parsed.APIURL == "" {
		parsed.APIURL = githubAPI
	}
	if parsed.Executable == nil {
		parsed.Executable = os.Executable
	}
	if parsed.CurrentVersion == "" {
		parsed.CurrentVersion = version.Version
	}
	if parsed.CurrentLabel == "" {
		parsed.CurrentLabel = version.String()
	}
	if !parsed.TTYSet {
		parsed.TTY = isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
	}
	return runUpgradeWithOptions(parsed)
}

func runUpgradeWithOptions(o upgradeOptions) int {
	// Refuse a mutation before any network call when it cannot be confirmed.
	if !o.Check && !o.DryRun && !o.Yes {
		if o.JSON || !o.TTY {
			msg := "refusing to replace the binary without a terminal; re-run with --yes or use --check"
			if o.JSON {
				msg = "confirmation required for --json; re-run with --yes or use --check"
			}
			return o.emit(upgradeResult{
				Type:    "upgrade",
				Action:  "declined",
				Current: strings.TrimPrefix(o.CurrentVersion, "v"),
				Mutated: false,
				Error:   msg,
			}, 3)
		}
	}

	release, err := fetchReleaseForUpgrade(o)
	if err != nil {
		return o.fail(err)
	}
	plan, err := buildUpgradePlan(o, release)
	if err != nil {
		return o.fail(err)
	}

	switch {
	case o.Check:
		return o.emit(upgradeResult{
			Type:      "upgrade",
			Action:    "check",
			Current:   plan.Current,
			Latest:    plan.LatestTag,
			Asset:     plan.AssetName,
			Target:    plan.Target,
			Mutated:   false,
			Migration: plan.migrationURL(),
		}, 0)
	case o.DryRun:
		return o.emit(upgradeResult{
			Type:      "upgrade",
			Action:    "dry-run",
			Current:   plan.Current,
			Latest:    plan.LatestTag,
			Asset:     plan.AssetName,
			Target:    plan.Target,
			Mutated:   false,
			Migration: plan.migrationURL(),
		}, 0)
	}

	if !plan.UpdateAvailable && !o.Force {
		_ = saveUpdateCache(updateCache{CheckedAt: time.Now().UTC(), LatestVersion: plan.LatestTag})
		return o.emit(upgradeResult{
			Type:    "upgrade",
			Action:  "up-to-date",
			Current: plan.Current,
			Latest:  plan.LatestTag,
			Mutated: false,
		}, 0)
	}

	if !o.Yes {
		if !o.TTY {
			return o.emit(upgradeResult{
				Type:    "upgrade",
				Action:  "declined",
				Current: plan.Current,
				Latest:  plan.LatestTag,
				Mutated: false,
				Error:   "refusing to replace the binary without a terminal; re-run with --yes",
			}, 3)
		}
		if !confirmUpgrade(o, plan) {
			return o.emit(upgradeResult{
				Type:    "upgrade",
				Action:  "declined",
				Current: plan.Current,
				Latest:  plan.LatestTag,
				Mutated: false,
				Error:   "upgrade cancelled",
			}, 3)
		}
	}

	fmt.Fprintf(o.Err, "  thermal %s  upgrading to %s...\n", o.CurrentLabel, plan.LatestTag)
	if err := performUpgrade(plan, o); err != nil {
		return o.fail(err)
	}
	_ = saveUpdateCache(updateCache{CheckedAt: time.Now().UTC(), LatestVersion: plan.LatestTag})
	return o.emit(upgradeResult{
		Type:      "upgrade",
		Action:    "upgraded",
		Current:   plan.Current,
		Latest:    plan.LatestTag,
		Asset:     plan.AssetName,
		Target:    plan.Target,
		Mutated:   true,
		Migration: plan.migrationURL(),
	}, 0)
}

func fetchReleaseForUpgrade(o upgradeOptions) (githubRelease, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.APIURL, nil)
	if err != nil {
		return githubRelease{}, fmt.Errorf("cannot create request: %w", err)
	}
	req.Header.Set("User-Agent", "thermal-upgrade")
	resp, err := o.Client.Do(req)
	if err != nil {
		return githubRelease{}, fmt.Errorf("cannot reach release API: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return githubRelease{}, fmt.Errorf("release API returned %s", resp.Status)
	}
	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return githubRelease{}, fmt.Errorf("cannot parse release info: %w", err)
	}
	if release.TagName == "" {
		return githubRelease{}, fmt.Errorf("release API returned an empty tag")
	}
	return release, nil
}

func buildUpgradePlan(o upgradeOptions, release githubRelease) (upgradePlan, error) {
	current := strings.TrimPrefix(o.CurrentVersion, "v")
	latestTag := release.TagName
	latest := strings.TrimPrefix(latestTag, "v")

	assetName, assetURL, err := findAsset(release.Assets)
	if err != nil {
		return upgradePlan{}, fmt.Errorf("%v; download manually: %s", err, release.HTMLURL)
	}
	target, err := o.Executable()
	if err != nil {
		return upgradePlan{}, fmt.Errorf("cannot find current binary: %w", err)
	}
	if resolved, resolveErr := filepath.EvalSymlinks(target); resolveErr == nil {
		target = resolved
	}

	updateAvailable := current == "dev" || semverCompare(latest, current) > 0
	if o.Force {
		updateAvailable = true
	}
	return upgradePlan{
		Current:         current,
		CurrentLabel:    o.CurrentLabel,
		LatestTag:       latestTag,
		Latest:          latest,
		AssetName:       assetName,
		AssetURL:        assetURL,
		Target:          target,
		ReleaseURL:      release.HTMLURL,
		UpdateAvailable: updateAvailable,
		Breaking:        isBreakingUpgrade(current, latest),
	}, nil
}

func (p upgradePlan) migrationURL() string {
	if p.Breaking {
		return "https://jadmadi.net/project/thermal/migration"
	}
	return ""
}

func confirmUpgrade(o upgradeOptions, plan upgradePlan) bool {
	fmt.Fprintf(o.Err, "Upgrade thermal %s to %s? [y/N] ", o.CurrentLabel, plan.LatestTag)
	line, _ := bufio.NewReader(o.Stdin).ReadString('\n')
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes"
}

func performUpgrade(plan upgradePlan, o upgradeOptions) error {
	dir := filepath.Dir(plan.Target)
	probe, err := os.CreateTemp(dir, ".thermal-upgrade-probe-*")
	if err != nil {
		return fmt.Errorf("target directory is not writable: %s", dir)
	}
	_ = probe.Close()
	_ = os.Remove(probe.Name())

	tmpDir, err := os.MkdirTemp("", "thermal-upgrade-*")
	if err != nil {
		return fmt.Errorf("cannot create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, plan.AssetName)
	if err := downloadFileWith(o.DownloadClient, plan.AssetURL, archivePath); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	binaryPath := filepath.Join(tmpDir, "thermal")
	if err := extractBinary(archivePath, plan.AssetName, binaryPath); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}
	if err := os.Chmod(binaryPath, 0o755); err != nil {
		return fmt.Errorf("cannot chmod new binary: %w", err)
	}

	newPath := plan.Target + ".new"
	oldPath := plan.Target + ".old"
	defer os.Remove(newPath)

	if err := copyFile(binaryPath, newPath); err != nil {
		return fmt.Errorf("cannot write new binary: %w", err)
	}
	if err := os.Chmod(newPath, 0o755); err != nil {
		return fmt.Errorf("cannot chmod new binary: %w", err)
	}

	_ = os.Remove(oldPath)
	if err := os.Rename(plan.Target, oldPath); err != nil {
		return fmt.Errorf("cannot replace binary: %w", err)
	}
	if err := os.Rename(newPath, plan.Target); err != nil {
		_ = os.Rename(oldPath, plan.Target)
		return fmt.Errorf("cannot replace binary: %w", err)
	}
	_ = os.Remove(oldPath)
	return nil
}

func (o upgradeOptions) fail(err error) int {
	if o.JSON {
		return o.emit(upgradeResult{
			Type:    "upgrade",
			Action:  "error",
			Current: strings.TrimPrefix(o.CurrentVersion, "v"),
			Mutated: false,
			Error:   err.Error(),
		}, 1)
	}
	fmt.Fprintf(o.Err, "thermal: %v\n", err)
	return 1
}

func (o upgradeOptions) emit(r upgradeResult, code int) int {
	if o.JSON {
		b, err := marshalJSON(r)
		if err != nil {
			fmt.Fprintf(o.Err, "thermal: cannot write JSON: %v\n", err)
			return 1
		}
		fmt.Fprintln(o.Out, string(b))
		return code
	}

	switch r.Action {
	case "check", "dry-run":
		fmt.Fprintf(o.Out, "thermal %s → %s\n", r.Current, r.Latest)
		fmt.Fprintf(o.Out, "  asset:  %s\n", r.Asset)
		fmt.Fprintf(o.Out, "  target: %s\n", r.Target)
		fmt.Fprintf(o.Out, "  action: %s\n", r.Action)
	case "up-to-date":
		fmt.Fprintf(o.Out, "already up to date — %s\n", r.Latest)
	case "upgraded":
		printUpgradeSuccess(o.Out, r.Current, r.Latest)
	case "declined", "error":
		fmt.Fprintf(o.Err, "thermal: %s\n", r.Error)
	}
	return code
}
