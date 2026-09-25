// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
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

func newLocalRequest(method, path string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	req.Host = "127.0.0.1:8080"
	return req
}

func TestServer_SecurityHeaders(t *testing.T) {
	srv := newTestServer()
	handler := srv.Handler()

	paths := []string{"/", "/api/health", "/api/stats"}
	for _, p := range paths {
		req := newLocalRequest(http.MethodGet, p)
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

	req := newLocalRequest(http.MethodGet, "/")
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

	req := newLocalRequest(http.MethodGet, "/api/health")
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
			req := newLocalRequest(http.MethodGet, route)
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
				req := newLocalRequest(http.MethodGet, "/api/stats")
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
	req := newLocalRequest(http.MethodGet, "/api/stream").WithContext(ctx)
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
	reqREST := newLocalRequest(http.MethodGet, "/api/telemetry")
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
	reqSSE := newLocalRequest(http.MethodGet, "/api/stream").WithContext(ctx)
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
	reqTool := newLocalRequest(http.MethodGet, "/api/telemetry?tool=claude")
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
	reqNoEst := newLocalRequest(http.MethodGet, "/api/telemetry?no-estimate=true")
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
	reqBad := newLocalRequest(http.MethodGet, "/api/telemetry?chart=true")
	recBad := httptest.NewRecorder()
	handler.ServeHTTP(recBad, reqBad)
	if recBad.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for unsupported query param, got %d", recBad.Code)
	}
	if !strings.Contains(recBad.Body.String(), "unsupported query parameter") {
		t.Errorf("expected error message to mention unsupported param, got: %s", recBad.Body.String())
	}

	// 6. Unknown tool rejection: 400 Bad Request
	reqUnknown := newLocalRequest(http.MethodGet, "/api/telemetry?tool=nonexistent_ai")
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

func TestServer_HostValidation(t *testing.T) {
	var collectorCalls atomic.Int32
	srv := NewWithOptions(Options{Offline: true})
	srv.SetCollector(func(opts Options) (*TelemetryData, error) {
		collectorCalls.Add(1)
		return &TelemetryData{TotalTokens: 100}, nil
	})
	handler := srv.Handler()

	validAuthorities := []string{
		"127.0.0.1:8080",
		"127.0.0.1",
		"localhost:8080",
		"localhost",
		"sub.localhost:8080",
		"[::1]:8080",
		"[::1]",
		"::1",
	}

	for _, auth := range validAuthorities {
		t.Run("valid_"+auth, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/telemetry", nil)
			req.Host = auth
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("authority %q failed with status %d: %s", auth, rec.Code, rec.Body.String())
			}
		})
	}

	untrustedAuthorities := []string{
		"untrusted.example:8080",
		"untrusted.example",
		"attacker.com:8080",
		"127.0.0.1:9999",
		"localhost:8081",
		"[::1:8080",
		"127.0.0.1:invalid",
		"127.0.0.1:0",
		"127.0.0.1:70000",
		"",
	}

	for _, auth := range untrustedAuthorities {
		t.Run("untrusted_"+auth, func(t *testing.T) {
			callsBefore := collectorCalls.Load()
			req := httptest.NewRequest(http.MethodGet, "/api/telemetry", nil)
			req.Host = auth
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Fatalf("expected status 403 for authority %q, got %d", auth, rec.Code)
			}
			if !strings.Contains(rec.Body.String(), "forbidden: untrusted request authority") {
				t.Errorf("expected rejection message for authority %q, got %s", auth, rec.Body.String())
			}
			if callsAfter := collectorCalls.Load(); callsAfter != callsBefore {
				t.Errorf("untrusted authority %q triggered data loading (calls: %d -> %d)", auth, callsBefore, callsAfter)
			}
		})
	}

	// Test rejection on SSE route /api/stream as well
	t.Run("untrusted_sse_stream", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/stream", nil)
		req.Host = "untrusted.example:8080"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403 for SSE with untrusted authority, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "forbidden: untrusted request authority") {
			t.Errorf("expected rejection message for SSE, got %s", rec.Body.String())
		}
	})
}

func TestServer_OriginAndCrossSiteValidation(t *testing.T) {
	srv := newTestServer()
	handler := srv.Handler()

	tests := []struct {
		name         string
		origin       string
		secFetchSite string
		wantStatus   int
	}{
		{
			name:       "valid_same_origin_localhost",
			origin:     "http://localhost:8080",
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid_same_origin_127",
			origin:     "http://127.0.0.1:8080",
			wantStatus: http.StatusOK,
		},
		{
			name:       "foreign_origin_rejected",
			origin:     "http://evil.com",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "null_origin_rejected",
			origin:     "null",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "mismatched_origin_port_rejected",
			origin:     "http://localhost:9999",
			wantStatus: http.StatusForbidden,
		},
		{
			name:         "sec_fetch_site_cross_site_rejected",
			origin:       "http://localhost:8080",
			secFetchSite: "cross-site",
			wantStatus:   http.StatusForbidden,
		},
		{
			name:         "sec_fetch_site_same_origin_allowed",
			origin:       "http://localhost:8080",
			secFetchSite: "same-origin",
			wantStatus:   http.StatusOK,
		},
		{
			name:       "no_origin_allowed_for_direct_cli",
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/telemetry", nil)
			req.Host = "127.0.0.1:8080"
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			if tc.secFetchSite != "" {
				req.Header.Set("Sec-Fetch-Site", tc.secFetchSite)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestServer_ExplicitHostAndWildcard(t *testing.T) {
	// 1. Explicit custom host: --host custom.lan --port 9000
	srvCustom := NewWithOptions(Options{Host: "custom.lan", Port: 9000, Offline: true})
	srvCustom.SetCollector(func(opts Options) (*TelemetryData, error) {
		return &TelemetryData{TotalTokens: 50}, nil
	})
	handlerCustom := srvCustom.Handler()

	reqCustom := httptest.NewRequest(http.MethodGet, "/api/telemetry", nil)
	reqCustom.Host = "custom.lan:9000"
	recCustom := httptest.NewRecorder()
	handlerCustom.ServeHTTP(recCustom, reqCustom)
	if recCustom.Code != http.StatusOK {
		t.Fatalf("custom host expected 200, got %d: %s", recCustom.Code, recCustom.Body.String())
	}

	reqCustomLoopback := httptest.NewRequest(http.MethodGet, "/api/telemetry", nil)
	reqCustomLoopback.Host = "localhost:9000"
	recCustomLoopback := httptest.NewRecorder()
	handlerCustom.ServeHTTP(recCustomLoopback, reqCustomLoopback)
	if recCustomLoopback.Code != http.StatusOK {
		t.Fatalf("custom host loopback alias expected 200, got %d: %s", recCustomLoopback.Code, recCustomLoopback.Body.String())
	}

	reqCustomOther := httptest.NewRequest(http.MethodGet, "/api/telemetry", nil)
	reqCustomOther.Host = "other.lan:9000"
	recCustomOther := httptest.NewRecorder()
	handlerCustom.ServeHTTP(recCustomOther, reqCustomOther)
	if recCustomOther.Code != http.StatusForbidden {
		t.Fatalf("foreign host on custom server expected 403, got %d", recCustomOther.Code)
	}

	// 2. Wildcard bind: --host 0.0.0.0 --port 8080
	srvWildcard := NewWithOptions(Options{Host: "0.0.0.0", Port: 8080, Offline: true})
	srvWildcard.SetCollector(func(opts Options) (*TelemetryData, error) {
		return &TelemetryData{TotalTokens: 50}, nil
	})
	handlerWildcard := srvWildcard.Handler()

	wildcardAllowed := []string{
		"127.0.0.1:8080",
		"localhost:8080",
		"0.0.0.0:8080",
		"192.168.1.55:8080",
	}
	for _, auth := range wildcardAllowed {
		req := httptest.NewRequest(http.MethodGet, "/api/telemetry", nil)
		req.Host = auth
		rec := httptest.NewRecorder()
		handlerWildcard.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("wildcard expected 200 for %q, got %d: %s", auth, rec.Code, rec.Body.String())
		}
	}

	// Wildcard MUST reject arbitrary DNS names to prevent DNS rebinding
	reqWildcardDNS := httptest.NewRequest(http.MethodGet, "/api/telemetry", nil)
	reqWildcardDNS.Host = "attacker.com:8080"
	recWildcardDNS := httptest.NewRecorder()
	handlerWildcard.ServeHTTP(recWildcardDNS, reqWildcardDNS)
	if recWildcardDNS.Code != http.StatusForbidden {
		t.Fatalf("wildcard bind must reject arbitrary DNS names, got %d", recWildcardDNS.Code)
	}
}

func TestServer_ForwardedHeadersIgnored(t *testing.T) {
	srv := newTestServer()
	handler := srv.Handler()

	req1 := httptest.NewRequest(http.MethodGet, "/api/telemetry", nil)
	req1.Host = "evil.com:8080"
	req1.Header.Set("X-Forwarded-Host", "localhost:8080")
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusForbidden {
		t.Fatalf("server must not trust X-Forwarded-Host, got %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/telemetry", nil)
	req2.Host = "evil.com:8080"
	req2.Header.Set("Forwarded", "host=localhost:8080")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusForbidden {
		t.Fatalf("server must not trust Forwarded header, got %d", rec2.Code)
	}
}

func TestServer_LoopbackIntegration(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to bind loopback listener: %v", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	srv := NewWithOptions(Options{Host: "127.0.0.1", Port: port, Offline: true})
	srv.SetCollector(func(opts Options) (*TelemetryData, error) {
		return &TelemetryData{
			TotalTokens:   12345,
			CurrentStreak: 3,
		}, nil
	})

	httpServer := &http.Server{
		Handler: srv.Handler(),
	}
	defer httpServer.Close()

	go func() {
		_ = httpServer.Serve(listener)
	}()

	client := &http.Client{Timeout: 3 * time.Second}

	// 1. Valid loopback request
	validURL := fmt.Sprintf("http://127.0.0.1:%d/api/telemetry", port)
	respValid, err := client.Get(validURL)
	if err != nil {
		t.Fatalf("loopback request failed: %v", err)
	}
	defer respValid.Body.Close()

	if respValid.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 for loopback request, got %d", respValid.StatusCode)
	}
	bodyBytes, _ := io.ReadAll(respValid.Body)
	var data TelemetryData
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		t.Fatalf("failed decoding telemetry json: %v", err)
	}
	if data.TotalTokens != 12345 {
		t.Errorf("expected 12345 tokens, got %d", data.TotalTokens)
	}

	// 2. Untrusted authority request over real network
	reqUntrusted, err := http.NewRequest(http.MethodGet, validURL, nil)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	reqUntrusted.Host = "untrusted.example"
	respUntrusted, err := client.Do(reqUntrusted)
	if err != nil {
		t.Fatalf("untrusted request failed: %v", err)
	}
	defer respUntrusted.Body.Close()

	if respUntrusted.StatusCode != http.StatusForbidden {
		t.Fatalf("expected status 403 for untrusted host authority, got %d", respUntrusted.StatusCode)
	}
	untrustedBody, _ := io.ReadAll(respUntrusted.Body)
	if !strings.Contains(string(untrustedBody), "forbidden: untrusted request authority") {
		t.Errorf("expected rejection message, got %s", string(untrustedBody))
	}
}
