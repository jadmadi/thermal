// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/jadmadi/thermal/internal/loaders"
	"github.com/jadmadi/thermal/internal/pricing"
	"github.com/jadmadi/thermal/internal/thermal"
	"github.com/jadmadi/thermal/internal/version"
)

//go:embed assets/index.html
var embeddedAssets embed.FS

// TelemetryData encapsulates the complete local telemetry snapshot for the web UI.
type TelemetryData struct {
	GeneratedAt   string                          `json:"generatedAt"`
	CurrentStreak int                             `json:"currentStreak"`
	LongestStreak int                             `json:"longestStreak"`
	ActiveDays    int                             `json:"activeDays"`
	TotalTokens   int64                           `json:"totalTokens"`
	InputTokens   int64                           `json:"inputTokens"`
	OutputTokens  int64                           `json:"outputTokens"`
	Reasoning     int64                           `json:"reasoningTokens"`
	CacheTokens   int64                           `json:"cacheTokens"`
	TotalCost     float64                         `json:"totalCost"`
	DailyActivity map[string]thermal.DayActivity  `json:"dailyActivity"`
	Results       []thermal.ToolResult            `json:"results"`
	Projects      []thermal.ProjectRow            `json:"projects"`
	Models        map[string]int64                `json:"models"`
}

// Server provides the local HTTP dashboard and telemetry API server.
type Server struct {
	host       string
	port       int
	offline    bool
	httpServer *http.Server
	mux        *http.ServeMux
	mu         sync.RWMutex
	cache      *TelemetryData
	cachedAt   time.Time
}

// New constructs a new Server instance.
func New(host string, port int, offline bool) *Server {
	if host == "" {
		host = "127.0.0.1"
	}
	if port <= 0 {
		port = 8080
	}

	s := &Server{
		host:    host,
		port:    port,
		offline: offline,
		mux:     http.NewServeMux(),
	}

	s.routes()
	return s
}

func (s *Server) securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) routes() {
	// Serve embedded HTML root
	s.mux.HandleFunc("/", s.handleIndex)

	// API routes
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/telemetry", s.handleTelemetry)
	s.mux.HandleFunc("/api/leaderboard", s.handleLeaderboard)
	s.mux.HandleFunc("/api/daily", s.handleDaily)
	s.mux.HandleFunc("/api/stats", s.handleStats)
	s.mux.HandleFunc("/api/projects", s.handleProjects)
	s.mux.HandleFunc("/api/models", s.handleModels)
}

// Handler returns the HTTP handler with all security headers and routing attached.
func (s *Server) Handler() http.Handler {
	return s.securityMiddleware(s.mux)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data, err := fs.ReadFile(embeddedAssets, "assets/index.html")
	if err != nil {
		http.Error(w, "Asset not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"status":    "ok",
		"version":   version.String(),
		"commit":    version.Commit,
		"cleanRoom": true,
		"time":      time.Now().UTC().Format(time.RFC3339),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) getOrFetchTelemetry() (*TelemetryData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cache != nil && time.Since(s.cachedAt) < 15*time.Second {
		return s.cache, nil
	}

	data, err := CollectTelemetry(s.offline)
	if err != nil {
		return nil, err
	}

	s.cache = data
	s.cachedAt = time.Now()
	return data, nil
}

func (s *Server) handleTelemetry(w http.ResponseWriter, r *http.Request) {
	data, err := s.getOrFetchTelemetry()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed collecting telemetry: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	data, err := s.getOrFetchTelemetry()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, data.Results)
}

func (s *Server) handleDaily(w http.ResponseWriter, r *http.Request) {
	data, err := s.getOrFetchTelemetry()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, data.DailyActivity)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	data, err := s.getOrFetchTelemetry()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	stats := map[string]interface{}{
		"currentStreak":   data.CurrentStreak,
		"longestStreak":   data.LongestStreak,
		"activeDays":      data.ActiveDays,
		"totalTokens":     data.TotalTokens,
		"inputTokens":     data.InputTokens,
		"outputTokens":    data.OutputTokens,
		"reasoningTokens": data.Reasoning,
		"cacheTokens":     data.CacheTokens,
		"totalCost":       data.TotalCost,
		"toolsCount":      len(data.Results),
	}
	writeJSON(w, http.StatusOK, stats)
}

func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	data, err := s.getOrFetchTelemetry()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, data.Projects)
}

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	data, err := s.getOrFetchTelemetry()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, data.Models)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// CollectTelemetry reads local tool data, calculates streaks, and aggregates metrics.
func CollectTelemetry(offline bool) (*TelemetryData, error) {
	allTools := loaders.AllTools()
	pricer := pricing.Load(pricing.DefaultCachePath(), offline)

	allDays := make(map[string]*thermal.DailyRow)
	var allProjects []thermal.ProjectDay
	allModels := make(map[string]int64)
	dailyActivity := make(map[string]thermal.DayActivity)
	activeDaysSet := make(map[string]bool)

	var results []thermal.ToolResult

	for t, info := range allTools {
		var hasData bool
		switch t {
		case thermal.ToolMiMoCode, thermal.ToolOpenCode, thermal.ToolDevin, thermal.ToolZCode, thermal.ToolMuse, thermal.ToolHermes:
			if info.DBPath != "" {
				if _, err := os.Stat(info.DBPath); err == nil {
					hasData = true
				}
			}
		default:
			if info.DataDir != "" {
				if _, err := os.Stat(info.DataDir); err == nil {
					hasData = true
				}
			}
		}

		if !hasData {
			continue
		}

		data, err := loaders.LoadToolData(t, info, "")
		if err != nil {
			continue
		}

		var estCost float64
		if pricer != nil && data.Summary.Cost == 0 {
			for _, d := range data.Daily {
				if len(d.Models) > 0 {
					c, _ := pricer.PriceDay(d)
					estCost += c
				}
			}
		}

		toolActiveDays := make(map[string]bool)
		for _, d := range data.Daily {
			if d.Turns > 0 {
				toolActiveDays[d.Day] = true
				activeDaysSet[d.Day] = true

				curAct := dailyActivity[d.Day]
				curAct.Tokens += d.Tokens
				curAct.Turns += d.Turns
				dailyActivity[d.Day] = curAct
			}

			// Aggregate into allDays
			existing := allDays[d.Day]
			if existing == nil {
				clone := d
				clone.Models = make(map[string]thermal.ModelTokens)
				for k, v := range d.Models {
					clone.Models[k] = v
				}
				allDays[d.Day] = &clone
			} else {
				existing.Tokens += d.Tokens
				existing.Turns += d.Turns
				existing.Input += d.Input
				existing.Output += d.Output
				existing.Reasoning += d.Reasoning
				existing.Cache += d.Cache
				existing.Cost += d.Cost
				for k, v := range d.Models {
					existing.Models[k] = existing.Models[k].Add(v)
				}
			}

			for m, mt := range d.Models {
				allModels[m] += mt.Total()
			}
		}

		current, longest := thermal.ComputeStreaks(toolActiveDays)
		cost := data.Summary.Cost
		if cost == 0 {
			cost = estCost
		}

		results = append(results, thermal.ToolResult{
			Tool:          t,
			Name:          info.Name,
			Summary:       data.Summary,
			Daily:         data.Daily,
			CurrentStreak: current,
			LongestStreak: longest,
			ActiveDays:    len(toolActiveDays),
			TotalActivity: data.Summary.LifetimeTokens,
			DataPath:      data.Path,
			EstimatedCost: estCost,
		})

		allProjects = append(allProjects, data.Projects...)
	}

	// Sort results by tokens desc
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalActivity > results[j].TotalActivity
	})

	// Aggregate projects
	projReport := thermal.AggregateProjects(allProjects, thermal.ProjectOptions{}, pricer)

	curStreak, longStreak := thermal.ComputeStreaks(activeDaysSet)

	var totTokens, inTokens, outTokens, reasTokens, cacheTokens int64
	var totCost float64
	for _, r := range results {
		totTokens += r.Summary.LifetimeTokens
		inTokens += r.Summary.InputTokens
		outTokens += r.Summary.OutputTokens
		reasTokens += r.Summary.ReasoningTokens
		cacheTokens += r.Summary.CacheTokens
		if r.Summary.Cost > 0 {
			totCost += r.Summary.Cost
		} else {
			totCost += r.EstimatedCost
		}
	}

	return &TelemetryData{
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		CurrentStreak: curStreak,
		LongestStreak: longStreak,
		ActiveDays:    len(activeDaysSet),
		TotalTokens:   totTokens,
		InputTokens:   inTokens,
		OutputTokens:  outTokens,
		Reasoning:     reasTokens,
		CacheTokens:   cacheTokens,
		TotalCost:     totCost,
		DailyActivity: dailyActivity,
		Results:       results,
		Projects:      projReport.Rows,
		Models:        allModels,
	}, nil
}

// StartServer starts the HTTP server listening on the configured host/port.
func StartServer(host string, port int, offline bool, openBrowser bool) error {
	s := New(host, port, offline)
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to bind %s: %w", addr, err)
	}

	s.httpServer = &http.Server{
		Handler:      s.Handler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	url := fmt.Sprintf("http://%s", addr)
	fmt.Printf("\n\033[1;36m═════════════════════════════════════════════════════════════════════\033[0m\n")
	fmt.Printf("\033[1m  THERMAL · Local Web Dashboard\033[0m\n")
	fmt.Printf("  Serving on \033[1;32m%s\033[0m\n", url)
	fmt.Printf("  Localhost security boundary · Zero cloud telemetry\n")
	fmt.Printf("  Press \033[1mCtrl+C\033[0m to terminate\n")
	fmt.Printf("\033[1;36m═════════════════════════════════════════════════════════════════════\033[0m\n\n")

	if openBrowser {
		go openBrowserURL(url)
	}

	// Trap termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case <-stop:
		fmt.Println("\nShutting down Thermal dashboard...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(ctx)
	case err := <-serverErr:
		return err
	}
}

func openBrowserURL(url string) {
	time.Sleep(200 * time.Millisecond)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
