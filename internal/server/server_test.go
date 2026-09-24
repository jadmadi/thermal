// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jadmadi/thermal/internal/loaders"
	"github.com/jadmadi/thermal/internal/thermal"
)

type mockPricer struct{}

func (m *mockPricer) PriceDay(day thermal.DailyRow) (float64, []string) {
	for model, mt := range day.Models {
		if model == "claude-3-5-sonnet" {
			return float64(mt.Total()) * 0.015, nil
		}
	}
	return 0, nil
}

func newTestServer() *Server {
	srv := NewWithOptions(Options{Offline: true})
	srv.SetCollector(func(opts Options) (*TelemetryData, error) {
		res := &TelemetryData{
			GeneratedAt:   "2026-09-24T12:00:00Z",
			CurrentStreak: 5,
			LongestStreak: 10,
			ActiveDays:    5,
			TotalTokens:   15000,
			InputTokens:   10000,
			OutputTokens:  5000,
			TotalCost:     1.25,
			RecordedCost:  1.00,
			EstimatedCost: 0.25,
			Results: []thermal.ToolResult{
				{
					Tool:          thermal.ToolClaude,
					Name:          "Claude",
					TotalActivity: 15000,
					Summary:       thermal.Summary{LifetimeTokens: 15000, Cost: 1.25},
				},
			},
			DailyActivity: map[string]thermal.DayActivity{
				"2026-09-24": {Tokens: 15000, Turns: 5},
			},
			Projects: []thermal.ProjectRow{
				{Project: "test-repo", Tokens: 15000, Cost: 1.25},
			},
			Models: map[string]int64{
				"claude-3-5-sonnet": 15000,
			},
		}

		if opts.Tool != "" && opts.Tool != "all" && opts.Tool != "auto" {
			if opts.Tool == "claude" {
				res.TotalTokens = 8000
			}
		}
		if opts.NoEstimate {
			res.EstimatedCost = 0
			res.TotalCost = res.RecordedCost
		}
		return res, nil
	})
	return srv
}

func TestServer_SecurityHeaders(t *testing.T) {
	srv := newTestServer()
	handler := srv.Handler()

	paths := []string{"/", "/api/health", "/api/stats"}
	for _, p := range paths {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Errorf("path %s missing X-Content-Type-Options: nosniff", p)
		}
		if rec.Header().Get("X-Frame-Options") != "DENY" {
			t.Errorf("path %s missing X-Frame-Options: DENY", p)
		}
		if rec.Header().Get("Referrer-Policy") != "no-referrer" {
			t.Errorf("path %s missing Referrer-Policy: no-referrer", p)
		}
		csp := rec.Header().Get("Content-Security-Policy")
		if !strings.Contains(csp, "default-src 'self'") {
			t.Errorf("path %s missing strict CSP, got %s", p, csp)
		}
	}
}

func TestServer_EmbeddedIndex(t *testing.T) {
	srv := newTestServer()
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("expected text/html content type, got %s", contentType)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "THERMAL") {
		t.Errorf("index body missing 'THERMAL'")
	}
	if !strings.Contains(body, "52-Week Contribution Activity Heatmap") {
		t.Errorf("index body missing heatmap section")
	}
}

func TestServer_HealthEndpoint(t *testing.T) {
	srv := newTestServer()
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &data); err != nil {
		t.Fatalf("failed decoding health json: %v", err)
	}
	if data["status"] != "ok" {
		t.Errorf("expected status ok, got %v", data["status"])
	}
	if data["cleanRoom"] != true {
		t.Errorf("expected cleanRoom true, got %v", data["cleanRoom"])
	}
}

func TestServer_APIRoutes(t *testing.T) {
	srv := newTestServer()
	handler := srv.Handler()

	routes := []string{
		"/api/telemetry",
		"/api/leaderboard",
		"/api/daily",
		"/api/stats",
		"/api/projects",
		"/api/models",
	}

	for _, route := range routes {
		t.Run(route, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, route, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("route %s returned status %d: %s", route, rec.Code, rec.Body.String())
			}
			contentType := rec.Header().Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				t.Errorf("route %s expected application/json, got %s", route, contentType)
			}
		})
	}
}

func TestServer_ConcurrentRequests(t *testing.T) {
	srv := newTestServer()
	handler := srv.Handler()

	var wg sync.WaitGroup
	workers := 10
	requestsPerWorker := 5

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < requestsPerWorker; i++ {
				req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				if rec.Code != http.StatusOK {
					t.Errorf("concurrent request failed with status %d", rec.Code)
				}
			}
		}()
	}
	wg.Wait()
}

func TestServer_Stream(t *testing.T) {
	srv := newTestServer()
	handler := srv.Handler()

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/api/stream", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	handler.ServeHTTP(rec, req)

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		t.Errorf("expected text/event-stream, got %s", contentType)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "event: telemetry") {
		t.Errorf("expected body to contain telemetry event, got %s", body)
	}
}

func TestServer_RESTandSSEParity(t *testing.T) {
	srv := newTestServer()
	handler := srv.Handler()

	// 1. REST GET /api/telemetry
	reqREST := httptest.NewRequest(http.MethodGet, "/api/telemetry", nil)
	recREST := httptest.NewRecorder()
	handler.ServeHTTP(recREST, reqREST)
	if recREST.Code != http.StatusOK {
		t.Fatalf("REST returned status %d", recREST.Code)
	}
	var restDoc map[string]interface{}
	if err := json.Unmarshal(recREST.Body.Bytes(), &restDoc); err != nil {
		t.Fatalf("failed unmarshaling REST JSON: %v", err)
	}

	// 2. SSE GET /api/stream
	ctx, cancel := context.WithCancel(context.Background())
	reqSSE := httptest.NewRequest(http.MethodGet, "/api/stream", nil).WithContext(ctx)
	recSSE := httptest.NewRecorder()
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	handler.ServeHTTP(recSSE, reqSSE)

	bodySSE := recSSE.Body.String()
	if !strings.Contains(bodySSE, "event: telemetry\ndata: ") {
		t.Fatalf("SSE missing event header: %s", bodySSE)
	}
	rawJSON := strings.TrimPrefix(bodySSE, "event: telemetry\ndata: ")
	idx := strings.Index(rawJSON, "\n\n")
	if idx > 0 {
		rawJSON = rawJSON[:idx]
	}
	var sseDoc map[string]interface{}
	if err := json.Unmarshal([]byte(rawJSON), &sseDoc); err != nil {
		t.Fatalf("failed unmarshaling SSE JSON: %v", err)
	}

	// Assert parity between REST and SSE snapshots
	if restDoc["totalTokens"] != sseDoc["totalTokens"] {
		t.Errorf("Parity mismatch on totalTokens: REST=%v, SSE=%v", restDoc["totalTokens"], sseDoc["totalTokens"])
	}
	if restDoc["totalCost"] != sseDoc["totalCost"] {
		t.Errorf("Parity mismatch on totalCost: REST=%v, SSE=%v", restDoc["totalCost"], sseDoc["totalCost"])
	}

	// 3. Supported query param filtering: tool
	reqTool := httptest.NewRequest(http.MethodGet, "/api/telemetry?tool=claude", nil)
	recTool := httptest.NewRecorder()
	handler.ServeHTTP(recTool, reqTool)
	if recTool.Code != http.StatusOK {
		t.Fatalf("tool filter returned %d", recTool.Code)
	}
	var toolDoc map[string]interface{}
	_ = json.Unmarshal(recTool.Body.Bytes(), &toolDoc)
	if toolDoc["totalTokens"].(float64) != 8000 {
		t.Errorf("tool filter expected 8000 tokens, got %v", toolDoc["totalTokens"])
	}

	// 4. Supported query param: no-estimate
	reqNoEst := httptest.NewRequest(http.MethodGet, "/api/telemetry?no-estimate=true", nil)
	recNoEst := httptest.NewRecorder()
	handler.ServeHTTP(recNoEst, reqNoEst)
	if recNoEst.Code != http.StatusOK {
		t.Fatalf("no-estimate returned %d", recNoEst.Code)
	}
	var noEstDoc map[string]interface{}
	_ = json.Unmarshal(recNoEst.Body.Bytes(), &noEstDoc)
	if val, ok := noEstDoc["estimatedCost"]; ok && val.(float64) != 0 {
		t.Errorf("no-estimate expected 0 or omitted estimatedCost, got %v", val)
	}

	// 5. Unsupported query param rejection: 400 Bad Request
	reqBad := httptest.NewRequest(http.MethodGet, "/api/telemetry?chart=true", nil)
	recBad := httptest.NewRecorder()
	handler.ServeHTTP(recBad, reqBad)
	if recBad.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for unsupported query param, got %d", recBad.Code)
	}
	if !strings.Contains(recBad.Body.String(), "unsupported query parameter") {
		t.Errorf("expected error message to mention unsupported param, got: %s", recBad.Body.String())
	}

	// 6. Unknown tool rejection: 400 Bad Request
	reqUnknown := httptest.NewRequest(http.MethodGet, "/api/telemetry?tool=nonexistent_ai", nil)
	recUnknown := httptest.NewRecorder()
	handler.ServeHTTP(recUnknown, reqUnknown)
	if recUnknown.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for unknown tool, got %d", recUnknown.Code)
	}
	if !strings.Contains(recUnknown.Body.String(), "unknown tool") {
		t.Errorf("expected error message to mention unknown tool, got: %s", recUnknown.Body.String())
	}
}

func TestCollectTelemetryWithDeps_Synthetic(t *testing.T) {
	claudeData := loaders.ToolData{
		Path: "/test/claude",
		Summary: thermal.Summary{
			LifetimeTokens: 150,
			Cost:           1.50,
		},
		Daily: []thermal.DailyRow{
			{
				Day:    "2026-09-20",
				Tokens: 100,
				Turns:  1,
				Input:  60,
				Output: 40,
				Cost:   1.50,
				Models: map[string]thermal.ModelTokens{
					"claude-3-5-sonnet": {Input: 60, Output: 40},
				},
			},
			{
				Day:    "2026-09-21",
				Tokens: 50,
				Turns:  1,
				Input:  30,
				Output: 20,
				Cost:   0,
				Models: map[string]thermal.ModelTokens{
					"claude-3-5-sonnet": {Input: 30, Output: 20},
				},
			},
		},
		Projects: []thermal.ProjectDay{
			{
				Project: "synthetic-repo",
				Day:     "2026-09-20",
				Tokens:  100,
				Cost:    1.50,
			},
			{
				Project: "synthetic-repo",
				Day:     "2026-09-21",
				Tokens:  50,
				Cost:    0,
			},
		},
	}

	agyData := loaders.ToolData{
		Path: "/test/agy",
		Summary: thermal.Summary{
			LifetimeTokens: 1,
			Cost:           0,
		},
		Daily: []thermal.DailyRow{
			{
				Day:    "2026-09-20",
				Tokens: 1,
				Turns:  1,
				Input:  0,
				Output: 0,
				Cost:   0,
			},
		},
	}

	mockLoader := func(t thermal.Tool, info loaders.ToolInfo, path string) (loaders.ToolData, error) {
		if t == thermal.ToolClaude {
			return claudeData, nil
		}
		if t == thermal.ToolAgy {
			return agyData, nil
		}
		return loaders.ToolData{}, nil
	}

	pricer := &mockPricer{}
	now, _ := time.Parse("2006-01-02", "2026-09-22")

	opts := Options{
		Now:        now,
		Offline:    true,
		NoEstimate: false,
	}

	data, err := CollectTelemetryWithDeps(opts, pricer, mockLoader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.TotalTokens != 150 {
		t.Errorf("TotalTokens = %d, want 150 (Agy 1-step excluded)", data.TotalTokens)
	}
	if data.TotalCost != 2.25 {
		t.Errorf("TotalCost = %v, want 2.25", data.TotalCost)
	}
	if data.RecordedCost != 1.50 {
		t.Errorf("RecordedCost = %v, want 1.50", data.RecordedCost)
	}
	if data.EstimatedCost != 0.75 {
		t.Errorf("EstimatedCost = %v, want 0.75", data.EstimatedCost)
	}
	if data.ActiveDays != 2 {
		t.Errorf("ActiveDays = %d, want 2", data.ActiveDays)
	}
	if len(data.Projects) == 0 || data.Projects[0].Project != "synthetic-repo" {
		t.Errorf("Projects = %+v, want synthetic-repo", data.Projects)
	}
}
