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
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/term"

	"github.com/jadmadi/thermal/internal/loaders"
	"github.com/jadmadi/thermal/internal/pricing"
	"github.com/jadmadi/thermal/internal/thermal"
	"github.com/jadmadi/thermal/internal/version"
)

//go:embed assets/index.html
var embeddedAssets embed.FS

// TelemetryData encapsulates the complete local telemetry snapshot for the web UI.
type TelemetryData = thermal.TelemetryData

// Options configures telemetry collection and server execution.
type Options struct {
	Host        string    `json:"host,omitempty"`
	Port        int       `json:"port,omitempty"`
	Tool        string    `json:"tool,omitempty"`
	Since       string    `json:"since,omitempty"`
	Until       string    `json:"until,omitempty"`
	Last        int       `json:"last,omitempty"`
	NoEstimate  bool      `json:"noEstimate,omitempty"`
	Offline     bool      `json:"offline,omitempty"`
	StartOfWeek string    `json:"startOfWeek,omitempty"`
	DBPath      string    `json:"dbPath,omitempty"`
	Verbose     bool      `json:"verbose,omitempty"`
	Now         time.Time `json:"-"`
}

type cacheEntry struct {
	data     *TelemetryData
	cachedAt time.Time
}

// Server provides the local HTTP dashboard and telemetry API server.
type Server struct {
	host       string
	port       int
	offline    bool
	options    Options
	httpServer *http.Server
	mux        *http.ServeMux
	mu         sync.RWMutex
	cache      map[string]*cacheEntry
	collector  func(opts Options) (*TelemetryData, error)
	pricer     thermal.Pricer
	clock      func() time.Time
}

// New constructs a new Server instance using legacy host, port, offline arguments.
func New(host string, port int, offline bool) *Server {
	return NewWithOptions(Options{
		Host:    host,
		Port:    port,
		Offline: offline,
	})
}

// NewWithOptions constructs a Server instance with rich options.
func NewWithOptions(opts Options) *Server {
	host := opts.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := opts.Port
	if port <= 0 {
		port = 8080
	}

	s := &Server{
		host:    host,
		port:    port,
		offline: opts.Offline,
		options: opts,
		mux:     http.NewServeMux(),
		cache:   make(map[string]*cacheEntry),
	}

	s.routes()
	return s
}

// SetCollector injects a custom telemetry collection function (useful in tests).
func (s *Server) SetCollector(fn func(opts Options) (*TelemetryData, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.collector = fn
}

// SetPricer injects a pricing engine.
func (s *Server) SetPricer(p thermal.Pricer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pricer = p
}

// SetClock injects a deterministic time source.
func (s *Server) SetClock(clk func() time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clock = clk
}

// parseAuthority splits authority (host or host:port) into a clean hostname/IP and optional port.
// It structurally handles IPv4, IPv6 (both bracketed "[::1]:8080" and bare "::1"), and named hosts.
func parseAuthority(authority string) (string, int, error) {
	if authority == "" {
		return "", 0, fmt.Errorf("empty host authority")
	}

	// Handle bracketed IPv6 addresses: e.g. [::1]:8080 or [::1]
	if strings.HasPrefix(authority, "[") {
		closeBracket := strings.Index(authority, "]")
		if closeBracket == -1 {
			return "", 0, fmt.Errorf("malformed IPv6 authority: missing closing bracket")
		}
		host := authority[1:closeBracket]
		rem := authority[closeBracket+1:]
		if rem == "" {
			return host, 0, nil
		}
		if !strings.HasPrefix(rem, ":") {
			return "", 0, fmt.Errorf("malformed authority trailing bracket: %s", rem)
		}
		portStr := rem[1:]
		port, err := strconv.Atoi(portStr)
		if err != nil || port <= 0 || port > 65535 {
			return "", 0, fmt.Errorf("invalid port in authority: %s", portStr)
		}
		return host, port, nil
	}

	// Handle addresses with colons
	if strings.Contains(authority, ":") {
		// Bare IPv6 address with multiple colons (e.g. "::1")
		if strings.Count(authority, ":") > 1 {
			if ip := net.ParseIP(authority); ip != nil {
				return authority, 0, nil
			}
			return "", 0, fmt.Errorf("malformed IPv6 authority: unbracketed with multiple colons")
		}

		h, pStr, err := net.SplitHostPort(authority)
		if err != nil {
			return "", 0, fmt.Errorf("invalid host authority: %w", err)
		}
		port, err := strconv.Atoi(pStr)
		if err != nil || port <= 0 || port > 65535 {
			return "", 0, fmt.Errorf("invalid port in authority: %s", pStr)
		}
		return h, port, nil
	}

	// Single host or IP with no port
	return authority, 0, nil
}

// isAllowedHost checks if the given host/IP matches an intentional local alias,
// loopback interface, or explicitly configured host. Wildcard binds (0.0.0.0 / ::)
// do NOT trust arbitrary DNS names to prevent DNS rebinding.
func (s *Server) isAllowedHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	host = strings.TrimPrefix(strings.TrimSuffix(host, "]"), "[")
	if host == "" {
		return false
	}

	// 1. Localhost names
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}

	s.mu.RLock()
	cfgHost := strings.ToLower(strings.TrimSpace(s.host))
	s.mu.RUnlock()

	// 2. IP address checks
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() {
			return true
		}
		if ip.IsUnspecified() {
			return true
		}
		if cfgHost != "" {
			if cfgIP := net.ParseIP(cfgHost); cfgIP != nil && cfgIP.Equal(ip) {
				return true
			}
		}
		// If bound to wildcard, allow any valid IP on the host (e.g. LAN access by IP)
		if cfgHost == "0.0.0.0" || cfgHost == "::" || cfgHost == "" {
			return true
		}
		return false
	}

	// 3. DNS names: wildcard bind must NEVER trust arbitrary DNS names
	if cfgHost == "0.0.0.0" || cfgHost == "::" || cfgHost == "" {
		return false
	}

	// Explicitly configured host (preserves explicit --host <name> behavior)
	cleanCfgHost := strings.TrimPrefix(strings.TrimSuffix(cfgHost, "]"), "[")
	return host == cleanCfgHost
}

// validateRequestAuthority enforces that incoming requests come from trusted local authorities.
// It protects against DNS rebinding and cross-origin telemetry exfiltration while allowing
// local browser dashboards, CLI tools, and explicitly configured hosts.
func (s *Server) validateRequestAuthority(r *http.Request) error {
	if r.Host == "" {
		return fmt.Errorf("missing host header")
	}

	reqHost, reqPort, err := parseAuthority(r.Host)
	if err != nil {
		return fmt.Errorf("malformed host authority: %w", err)
	}

	s.mu.RLock()
	configuredPort := s.port
	s.mu.RUnlock()

	// If a port is specified in the Host header, it must match the listening port
	if reqPort > 0 && reqPort != configuredPort {
		return fmt.Errorf("invalid port in authority: got %d, expected %d", reqPort, configuredPort)
	}

	if !s.isAllowedHost(reqHost) {
		return fmt.Errorf("untrusted host authority: %s", reqHost)
	}

	// Validate browser Origin header if present (same-origin policy enforcement)
	if origin := r.Header.Get("Origin"); origin != "" {
		if origin == "null" {
			return fmt.Errorf("forbidden origin: null")
		}
		u, err := url.Parse(origin)
		if err != nil {
			return fmt.Errorf("malformed origin: %w", err)
		}
		origHost, origPort, err := parseAuthority(u.Host)
		if err != nil {
			return fmt.Errorf("malformed origin host: %w", err)
		}
		if origPort > 0 && origPort != configuredPort {
			return fmt.Errorf("forbidden origin port: %d, expected %d", origPort, configuredPort)
		}
		if !s.isAllowedHost(origHost) {
			return fmt.Errorf("forbidden cross-origin host: %s", origHost)
		}
	}

	// Reject cross-site requests signaled by modern browsers
	if sfs := r.Header.Get("Sec-Fetch-Site"); sfs == "cross-site" {
		return fmt.Errorf("forbidden cross-site request: %s", sfs)
	}

	return nil
}

func (s *Server) securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")

		if err := s.validateRequestAuthority(r); err != nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"forbidden: untrusted request authority"}`))
			return
		}

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
	s.mux.HandleFunc("/api/stream", s.handleStream)
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

func (s *Server) parseRequestOptions(r *http.Request) (Options, error) {
	opts := s.options

	q := r.URL.Query()
	for param := range q {
		switch param {
		case "tool", "since", "until", "last", "no-estimate", "offline":
		default:
			return opts, fmt.Errorf("unsupported query parameter: %s", param)
		}
	}

	if t := q.Get("tool"); t != "" {
		resolved, ok := loaders.ResolveTool(t)
		if !ok && t != "all" && t != "auto" {
			return opts, fmt.Errorf("unknown tool: %s", t)
		}
		if ok {
			opts.Tool = string(resolved)
		} else {
			opts.Tool = t
		}
	}
	if sStr := q.Get("since"); sStr != "" {
		if _, ok := thermal.ParseDay(sStr); !ok {
			return opts, fmt.Errorf("invalid since date: %s (expected YYYY-MM-DD)", sStr)
		}
		opts.Since = sStr
	}
	if uStr := q.Get("until"); uStr != "" {
		if _, ok := thermal.ParseDay(uStr); !ok {
			return opts, fmt.Errorf("invalid until date: %s (expected YYYY-MM-DD)", uStr)
		}
		opts.Until = uStr
	}
	if lStr := q.Get("last"); lStr != "" {
		n, err := strconv.Atoi(lStr)
		if err != nil || n < 0 {
			return opts, fmt.Errorf("invalid last parameter: %s", lStr)
		}
		opts.Last = n
	}
	if opts.Last > 0 && (opts.Since != "" || opts.Until != "") {
		return opts, fmt.Errorf("--last cannot be combined with --since or --until")
	}
	if ne := q.Get("no-estimate"); ne == "true" || ne == "1" {
		opts.NoEstimate = true
	}
	if off := q.Get("offline"); off == "true" || off == "1" {
		opts.Offline = true
	}

	return opts, nil
}

func (s *Server) getOrFetchTelemetry(opts Options) (*TelemetryData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if s.clock != nil {
		now = s.clock()
	}

	key := fmt.Sprintf("%s|%s|%s|%d|%v|%v", opts.Tool, opts.Since, opts.Until, opts.Last, opts.NoEstimate, opts.Offline)
	if entry, ok := s.cache[key]; ok && now.Sub(entry.cachedAt) < 15*time.Second {
		return entry.data, nil
	}

	var data *TelemetryData
	var err error
	if s.collector != nil {
		data, err = s.collector(opts)
	} else {
		data, err = CollectTelemetryWithDeps(opts, s.pricer, nil)
	}
	if err != nil {
		return nil, err
	}

	s.cache[key] = &cacheEntry{
		data:     data,
		cachedAt: now,
	}
	return data, nil
}

func (s *Server) handleTelemetry(w http.ResponseWriter, r *http.Request) {
	opts, err := s.parseRequestOptions(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	data, err := s.getOrFetchTelemetry(opts)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	opts, err := s.parseRequestOptions(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	data, err := s.getOrFetchTelemetry(opts)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, data.Results)
}

func (s *Server) handleDaily(w http.ResponseWriter, r *http.Request) {
	opts, err := s.parseRequestOptions(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	data, err := s.getOrFetchTelemetry(opts)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, data.DailyActivity)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	opts, err := s.parseRequestOptions(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	data, err := s.getOrFetchTelemetry(opts)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
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
		"recordedCost":    data.RecordedCost,
		"estimatedCost":   data.EstimatedCost,
		"unpricedTokens":  data.UnpricedTokens,
		"toolsCount":      len(data.Results),
	}
	writeJSON(w, http.StatusOK, stats)
}

func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	opts, err := s.parseRequestOptions(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	data, err := s.getOrFetchTelemetry(opts)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, data.Projects)
}

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	opts, err := s.parseRequestOptions(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	data, err := s.getOrFetchTelemetry(opts)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, data.Models)
}

func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	opts, err := s.parseRequestOptions(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Send initial snapshot
	data, err := s.getOrFetchTelemetry(opts)
	if err == nil {
		b, _ := json.Marshal(data)
		fmt.Fprintf(w, "event: telemetry\ndata: %s\n\n", b)
		flusher.Flush()
	}

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			s.mu.Lock()
			s.cache = make(map[string]*cacheEntry)
			s.mu.Unlock()
			data, err := s.getOrFetchTelemetry(opts)
			if err == nil {
				b, _ := json.Marshal(data)
				fmt.Fprintf(w, "event: telemetry\ndata: %s\n\n", b)
				flusher.Flush()
			}
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// CollectTelemetry reads local tool data, calculates streaks, and aggregates metrics.
func CollectTelemetry(opts Options) (*TelemetryData, error) {
	return CollectTelemetryWithDeps(opts, nil, nil)
}

// CollectTelemetryWithDeps collects telemetry with optional custom pricer and loader injections.
func CollectTelemetryWithDeps(opts Options, pricer thermal.Pricer, customLoader func(t thermal.Tool, info loaders.ToolInfo, path string) (loaders.ToolData, error)) (*TelemetryData, error) {
	isCustomLoader := customLoader != nil
	if customLoader == nil {
		customLoader = loaders.LoadToolData
	}

	if opts.NoEstimate {
		pricer = nil
	} else if pricer == nil {
		pricer = pricing.Load(pricing.DefaultCachePath(), opts.Offline)
	}

	allTools := loaders.AllTools()
	targetTool := strings.ToLower(strings.TrimSpace(opts.Tool))

	var toolsToLoad []thermal.Tool
	if targetTool != "" && targetTool != "all" && targetTool != "auto" {
		t, ok := loaders.ResolveTool(targetTool)
		if !ok {
			return nil, fmt.Errorf("unknown tool: %s", opts.Tool)
		}
		toolsToLoad = append(toolsToLoad, t)
	} else {
		for t := range allTools {
			toolsToLoad = append(toolsToLoad, t)
		}
		sort.Slice(toolsToLoad, func(i, j int) bool {
			return string(toolsToLoad[i]) < string(toolsToLoad[j])
		})
	}

	var results []thermal.ToolResult
	var allProjects []thermal.ProjectDay

	for _, t := range toolsToLoad {
		info, exists := allTools[t]
		if !exists {
			continue
		}

		var hasData bool
		if isCustomLoader {
			hasData = true
		} else {
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

			if opts.DBPath != "" && len(toolsToLoad) == 1 {
				hasData = true
			}
		}

		if !hasData {
			continue
		}

		dataPath := ""
		if len(toolsToLoad) == 1 {
			dataPath = opts.DBPath
		}

		data, err := customLoader(t, info, dataPath)
		if err != nil {
			if len(toolsToLoad) == 1 {
				return nil, err
			}
			continue
		}

		if isCustomLoader && data.Summary.LifetimeTokens == 0 && data.Summary.Cost == 0 && len(data.Daily) == 0 && len(data.Projects) == 0 {
			continue
		}

		results = append(results, thermal.ToolResult{
			Tool:     t,
			Name:     info.Name,
			Summary:  data.Summary,
			Daily:    data.Daily,
			DataPath: data.Path,
		})

		for i := range data.Projects {
			data.Projects[i].Tool = info.Name
		}
		allProjects = append(allProjects, data.Projects...)
	}

	teleOpts := thermal.TelemetryOptions{
		Tool:        opts.Tool,
		Since:       opts.Since,
		Until:       opts.Until,
		Last:        opts.Last,
		Now:         opts.Now,
		NoEstimate:  opts.NoEstimate,
		Offline:     opts.Offline,
		StartOfWeek: opts.StartOfWeek,
	}

	return thermal.AggregateTelemetry(results, allProjects, teleOpts, pricer), nil
}

// StartServer starts the HTTP server listening on the configured host/port.
func StartServer(host string, port int, offline bool, openBrowser bool) error {
	return StartServerWithOptions(Options{
		Host:    host,
		Port:    port,
		Offline: offline,
	}, openBrowser)
}

// StartServerWithOptions starts the HTTP server with rich options.
func StartServerWithOptions(opts Options, openBrowser bool) error {
	s := NewWithOptions(opts)

	var listener net.Listener
	var err error
	actualPort := s.port

	maxAttempts := 1
	if s.port == 8080 {
		maxAttempts = 10
	}

	for i := 0; i < maxAttempts; i++ {
		tryPort := s.port + i
		tryAddr := fmt.Sprintf("%s:%d", s.host, tryPort)
		listener, err = net.Listen("tcp", tryAddr)
		if err == nil {
			actualPort = tryPort
			s.port = tryPort
			break
		}
	}

	if err != nil {
		addr := fmt.Sprintf("%s:%d", s.host, s.port)
		return fmt.Errorf("failed to bind %s: %w (try specifying an alternate port with --port <N>)", addr, err)
	}

	s.httpServer = &http.Server{
		Handler:      s.Handler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	url := fmt.Sprintf("http://%s", addr)
	fmt.Printf("\n\033[1;36m═════════════════════════════════════════════════════════════════════\033[0m\n")
	fmt.Printf("\033[1m  THERMAL · Local Web Dashboard\033[0m\n")
	if actualPort != opts.Port && opts.Port > 0 {
		fmt.Printf("  \033[33mNote: Port %d was in use; switched to %d\033[0m\n", opts.Port, actualPort)
	}
	fmt.Printf("  Serving on \033[1;32m%s\033[0m\n", url)
	fmt.Printf("  Localhost security boundary · Zero cloud telemetry\n")
	fmt.Printf("  Press \033[1mq\033[0m or \033[1mCtrl+C\033[0m to terminate\n")
	fmt.Printf("\033[1;36m═════════════════════════════════════════════════════════════════════\033[0m\n\n")

	if openBrowser {
		go openBrowserURL(url)
	}

	// Trap termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	quitKey := make(chan struct{})
	if term.IsTerminal(int(os.Stdin.Fd())) {
		if oldState, err := term.MakeRaw(int(os.Stdin.Fd())); err == nil {
			defer func() {
				_ = term.Restore(int(os.Stdin.Fd()), oldState)
			}()
			go func() {
				var buf [1]byte
				for {
					n, err := os.Stdin.Read(buf[:])
					if err != nil || n == 0 {
						return
					}
					// 'q', 'Q', Ctrl+C (0x03), or ESC (0x1b)
					if buf[0] == 'q' || buf[0] == 'Q' || buf[0] == 3 || buf[0] == 27 {
						select {
						case <-quitKey:
						default:
							close(quitKey)
						}
						return
					}
				}
			}()
		}
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case <-stop:
		fmt.Println("\r\nShutting down Thermal dashboard...")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(ctx)
	case <-quitKey:
		fmt.Println("\r\nQuitting Thermal dashboard...")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
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
