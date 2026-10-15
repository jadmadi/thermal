// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type streamResult struct {
	Interval       string  `json:"interval"`
	WallSeconds    float64 `json:"wallSeconds"`
	WarmupSeconds  float64 `json:"warmupSeconds"`
	CPUSeconds     float64 `json:"cpuSeconds"`
	CPUPercent     float64 `json:"cpuPercent"`
	Lines          int     `json:"refreshLines"`
	Events         int     `json:"changeEvents"`
	InvalidLines   int     `json:"invalidLines"`
	FreshnessMs    float64 `json:"freshnessMs"`
	FreshnessToks  int64   `json:"freshnessTokens"`
	CancellationMs float64 `json:"cancellationMs"`
	CPUSource      string  `json:"cpuSource"`
}

type streamLine struct {
	at          time.Time
	hasSnapshot bool
	tokens      int64
	hasEvent    bool
	eventTool   string
	eventTokens int64
}

// parseStreamLine extracts the snapshot counter or change event from one NDJSON
// line. A line is invalid only when it is not JSON at all.
func parseStreamLine(line []byte) (streamLine, bool) {
	var payload struct {
		Type     string `json:"type"`
		Snapshot *struct {
			TodayTokens int64 `json:"todayTokens"`
		} `json:"snapshot"`
		Event *struct {
			Tool   string `json:"tool"`
			Tokens int64  `json:"tokens"`
		} `json:"event"`
	}
	if err := json.Unmarshal(line, &payload); err != nil {
		return streamLine{}, false
	}
	out := streamLine{}
	if payload.Snapshot != nil {
		out.hasSnapshot = true
		out.tokens = payload.Snapshot.TodayTokens
	}
	if payload.Event != nil {
		out.hasEvent = true
		out.eventTool = payload.Event.Tool
		out.eventTokens = payload.Event.Tokens
	}
	return out, true
}

// measureStream runs the live NDJSON monitor for a warm-up period plus the
// measured window, appends one source change inside the window, then cancels
// the process. A stream that ends on its own, emits a malformed line, or never
// reflects the appended source fails the measurement.
func measureStream(bin string, env []string, interval, warmup, window time.Duration, appendSource func() (int64, error)) (streamResult, error) {
	cmd := exec.Command(bin, "live", "--json", "--stream", "--interval", interval.String(), "--offline")
	cmd.Env = env
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return streamResult{}, err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	start := time.Now()
	if err := cmd.Start(); err != nil {
		return streamResult{}, err
	}

	lines := make(chan streamLine, 8192)
	readDone := make(chan error, 1)
	go func() {
		defer close(lines)
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 64*1024), 8*1024*1024)
		for scanner.Scan() {
			line, valid := parseStreamLine(scanner.Bytes())
			if !valid {
				readDone <- fmt.Errorf("malformed NDJSON line: %.120q", scanner.Text())
				return
			}
			line.at = time.Now()
			lines <- line
		}
		readDone <- scanner.Err()
	}()

	var (
		refreshes     int
		events        int
		freshness     time.Duration
		freshnessToks int64
		appendAt      time.Time
		appended      bool
	)

	phase := time.NewTimer(warmup + window)
	defer phase.Stop()
	// Change the source a third of the way into the window with a deliberate
	// sub-second offset so the sample does not land on a poll-tick boundary.
	appendTimer := time.NewTimer(warmup + window/3 + 137*time.Millisecond)
	defer appendTimer.Stop()

	readDoneReceived := false
	var earlyExit error

	// Phase 1: warm-up plus measured window. The stream emits the initial
	// snapshot once and a change event per detected source delta.
loop:
	for {
		select {
		case line, ok := <-lines:
			if !ok {
				if earlyExit == nil {
					earlyExit = errors.New("live stream stdout closed without cancellation")
				}
				break loop
			}
			refreshes++
			if line.hasEvent {
				events++
				if appended && freshness == 0 && line.eventTool == "Claude" && line.eventTokens > 0 {
					freshness = line.at.Sub(appendAt)
					freshnessToks = line.eventTokens
				}
			}
		case readErr := <-readDone:
			readDoneReceived = true
			if readErr != nil {
				earlyExit = readErr
			} else {
				earlyExit = errors.New("live stream stdout closed without cancellation")
			}
			break loop
		case <-appendTimer.C:
			toks, aerr := appendSource()
			if aerr != nil {
				earlyExit = fmt.Errorf("append source: %w", aerr)
				break loop
			}
			appended = true
			appendAt = time.Now()
			_ = toks
		case <-phase.C:
			break loop
		}
	}

	if earlyExit != nil {
		_ = cmd.Process.Kill()
		if !readDoneReceived {
			<-readDone
		}
		_ = cmd.Wait()
		return streamResult{}, fmt.Errorf("live stream ended unexpectedly: %w (stderr: %s)", earlyExit, stderr.String())
	}

	// Phase 2: deliberate termination. A healthy stream accepts this timeout;
	// any other early exit was already rejected above.
	cancelStart := time.Now()
	if err := signalInterrupt(cmd.Process); err != nil {
		_ = cmd.Process.Kill()
	}
	if !readDoneReceived {
		select {
		case readErr := <-readDone:
			if readErr != nil && !errors.Is(readErr, os.ErrClosed) {
				// A malformed line discovered while draining still fails the run.
				_ = cmd.Wait()
				return streamResult{}, fmt.Errorf("live stream read error at cancellation: %w", readErr)
			}
		case <-time.After(10 * time.Second):
			_ = cmd.Process.Kill()
			<-readDone
		}
	}
	for line := range lines {
		refreshes++
		if line.hasEvent {
			events++
		}
	}
	cancellation := time.Since(cancelStart)
	_ = cmd.Wait()

	cpuSeconds, cpuSource, cpuOK := processCPU(cmd.ProcessState)
	result := streamResult{
		Interval:       intervalLabel(interval),
		WallSeconds:    time.Since(start).Seconds(),
		WarmupSeconds:  warmup.Seconds(),
		Lines:          refreshes,
		Events:         events,
		FreshnessMs:    float64(freshness.Microseconds()) / 1000.0,
		FreshnessToks:  freshnessToks,
		CancellationMs: float64(cancellation.Microseconds()) / 1000.0,
		CPUSource:      cpuSource,
	}
	if cpuOK {
		result.CPUSeconds = cpuSeconds.Seconds()
		if wall := result.WallSeconds; wall > 0 {
			result.CPUPercent = result.CPUSeconds / wall * 100
		}
	}
	if result.Lines == 0 {
		return streamResult{}, fmt.Errorf("live stream produced zero NDJSON lines in %s", intervalLabel(interval))
	}
	if !appended {
		return streamResult{}, errors.New("source change was never appended; freshness unmeasured")
	}
	if result.FreshnessMs <= 0 {
		return streamResult{}, errors.New("source change was never reflected in the live stream")
	}
	return result, nil
}

func intervalLabel(d time.Duration) string {
	if d == time.Second {
		return "1s"
	}
	return d.String()
}

func signalInterrupt(p *os.Process) error {
	if err := p.Signal(os.Interrupt); err != nil {
		return p.Kill()
	}
	return nil
}
