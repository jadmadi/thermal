// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func streamFixtureHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	dir := filepath.Join(home, ".claude", "projects", "stream-test")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	line := fmt.Sprintf(`{"type":"assistant","timestamp":%q,"cwd":%q,"message":{"id":"stream-test","model":"claude-3-5-sonnet","usage":{"input_tokens":1000,"output_tokens":500,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}}`+"\n", now, home)
	if err := os.WriteFile(filepath.Join(dir, "session.jsonl"), []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}
	return home
}

func streamEnv(home string) []string {
	env := append([]string{}, os.Environ()...)
	for _, key := range []string{"GROK_HOME", "DSH_HOME", "HERMES_HOME", "CODEX_HOME", "OPENCODE_HOME", "ZED_HOME", "AGY_HOME"} {
		env = append(env, key+"=")
	}
	env = append(env, "HOME="+home, "XDG_DATA_HOME="+filepath.Join(home, ".local", "share"))
	return env
}

func TestLiveStreamClosesOnBrokenPipe(t *testing.T) {
	home := streamFixtureHome(t)
	bin := filepath.Join(t.TempDir(), "thermal-stream-test")
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}

	cmd := exec.Command(bin, "live", "--json", "--stream", "--interval", "100ms", "--offline")
	cmd.Env = streamEnv(home)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}()

	type readResult struct {
		line string
		err  error
	}
	readCh := make(chan readResult, 1)
	go func() {
		line, err := bufio.NewReader(stdout).ReadString('\n')
		readCh <- readResult{line: line, err: err}
	}()

	var first string
	select {
	case r := <-readCh:
		if r.err != nil {
			t.Fatalf("read first line: %v; stderr=%s", r.err, stderr.String())
		}
		first = r.line
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for the first snapshot; stderr=%s", stderr.String())
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(first), &payload); err != nil {
		t.Fatalf("first line is not JSON: %v\n%s", err, first)
	}
	if payload["type"] != "snapshot" || payload["snapshot"] == nil {
		t.Fatalf("unexpected first payload: %s", first)
	}
	if !strings.Contains(first, `"type":"snapshot"`) {
		t.Fatalf("first line is not compact NDJSON: %s", first)
	}

	// Close the consumer; the next periodic snapshot write must terminate the
	// stream cleanly rather than hanging.
	_ = stdout.Close()
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("stream did not exit cleanly after closed pipe: %v; stderr=%s", err, stderr.String())
		}
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatalf("stream hung after the consumer closed; stderr=%s", stderr.String())
	}
}
