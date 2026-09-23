// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestServer_SecurityHeaders(t *testing.T) {
	srv := New("127.0.0.1", 8080, true)
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
	srv := New("127.0.0.1", 8080, true)
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
	srv := New("127.0.0.1", 8080, true)
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
	srv := New("127.0.0.1", 8080, true)
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
	srv := New("127.0.0.1", 8080, true)
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
