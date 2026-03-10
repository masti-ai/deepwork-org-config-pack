// Package server implements the gt-proxy-server with mTLS support
package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Deepwork-AI/ai-planogram/proxy-server/pkg/allowlist"
	"github.com/Deepwork-AI/ai-planogram/proxy-server/pkg/audit"
	"github.com/Deepwork-AI/ai-planogram/proxy-server/pkg/ratelimit"
	"github.com/Deepwork-AI/ai-planogram/proxy-server/pkg/security"

	"golang.org/x/time/rate"
)

// Server represents the gt-proxy-server instance
type Server struct {
	addr        string
	caCertPath  string
	caKeyPath   string
	certDir     string
	rateLimiter *ratelimit.Limiter
	clients     map[string]*ClientInfo
	clientsMu   sync.RWMutex
	allowlist   *allowlist.Allowlist
	logger      *log.Logger
	auditLogger *audit.Logger
	sanitizer   *security.Sanitizer
	config      Config
}

// ClientInfo stores information about a connected client
type ClientInfo struct {
	Name          string
	CertHash      string
	LastSeen      time.Time
	RequestCount  int64
	RateLimiter   *rate.Limiter
	AllowedScopes []string // Git branches, directories, etc.
}

// Config holds server configuration
type Config struct {
	Addr       string
	CACertPath string
	CAKeyPath  string
	CertDir    string
	RateLimit  ratelimit.Config
	Allowlist  *allowlist.Allowlist
	AuditConfig audit.Config
	SecurityConfig security.Config
}

// New creates a new proxy server
func New(cfg Config) (*Server, error) {
	if cfg.Addr == "" {
		cfg.Addr = ":9876"
	}
	if cfg.CACertPath == "" {
		return nil, fmt.Errorf("CA certificate path is required")
	}
	if cfg.CAKeyPath == "" {
		return nil, fmt.Errorf("CA key path is required")
	}
	if cfg.CertDir == "" {
		cfg.CertDir = filepath.Join(os.Getenv("HOME"), "gt", ".runtime", "certs")
	}

	// Ensure cert directory exists
	if err := os.MkdirAll(cfg.CertDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create cert directory: %w", err)
	}

	// Create audit logger
	auditLogger, err := audit.New(cfg.AuditConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit logger: %w", err)
	}

	// Create security sanitizer
	sanitizer, err := security.NewSanitizer(cfg.SecurityConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create security sanitizer: %w", err)
	}

	s := &Server{
		addr:        cfg.Addr,
		caCertPath:  cfg.CACertPath,
		caKeyPath:   cfg.CAKeyPath,
		certDir:     cfg.CertDir,
		clients:     make(map[string]*ClientInfo),
		allowlist:   cfg.Allowlist,
		logger:      log.New(os.Stdout, "[proxy-server] ", log.LstdFlags),
		auditLogger: auditLogger,
		sanitizer:   sanitizer,
		config:      cfg,
		rateLimiter: ratelimit.New(cfg.RateLimit),
	}

	return s, nil
}

// Start starts the proxy server
func (s *Server) Start() error {
	// Load CA certificate for client verification
	caCert, err := os.ReadFile(s.caCertPath)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return fmt.Errorf("failed to parse CA certificate")
	}

	// Configure TLS
	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
		MinVersion: tls.VersionTLS13,
	}

	// Create HTTP server with security middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/exec", s.handleExec)
	mux.HandleFunc("/git/", s.handleGit)
	mux.HandleFunc("/gt/", s.handleGT)
	mux.HandleFunc("/client/register", s.handleClientRegister)

	// Chain middleware: security -> rate limiting -> handlers
	handler := s.middleware(mux)
	handler = security.SecurityMiddleware(s.config.SecurityConfig)(handler)
	handler = security.BodySizeLimitMiddleware(s.sanitizer.MaxBodySize())(handler)

	server := &http.Server{
		Addr:         s.addr,
		Handler:      handler,
		TLSConfig:    tlsConfig,
		ReadTimeout:  s.sanitizer.RequestTimeout() + 5*time.Second,
		WriteTimeout: s.sanitizer.RequestTimeout() + 10*time.Second,
		IdleTimeout:  120 * time.Second,
	}

	s.logger.Printf("Starting gt-proxy-server on %s", s.addr)
	s.logger.Printf("CA certificate: %s", s.caCertPath)
	s.logger.Printf("Client certificates directory: %s", s.certDir)
	s.logger.Printf("Max request body size: %d bytes", s.sanitizer.MaxBodySize())
	s.logger.Printf("Request timeout: %v", s.sanitizer.RequestTimeout())

	return server.ListenAndServeTLS(s.caCertPath, s.caKeyPath)
}

// middleware applies common middleware (logging, rate limiting, audit)
func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Extract client certificate info
		clientName := s.getClientName(r)
		if clientName == "" {
			s.auditLogger.LogRequest("unknown", r.Method, r.URL.Path, false, http.StatusUnauthorized, time.Since(start))
			http.Error(w, "Client certificate required", http.StatusUnauthorized)
			return
		}

		// Determine endpoint type for rate limiting
		endpoint := ratelimit.EndpointAPI
		if strings.HasPrefix(r.URL.Path, "/exec") {
			endpoint = ratelimit.EndpointExec
		} else if strings.HasPrefix(r.URL.Path, "/git") {
			endpoint = ratelimit.EndpointGit
		} else if strings.HasPrefix(r.URL.Path, "/gt") {
			endpoint = ratelimit.EndpointGT
		}

		// Apply rate limiting
		if !s.rateLimiter.Allow(clientName, endpoint) {
			s.logger.Printf("Rate limit exceeded for client: %s", clientName)
			s.auditLogger.LogRequest(clientName, r.Method, r.URL.Path, false, http.StatusTooManyRequests, time.Since(start))
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Check concurrent connections
		if !s.rateLimiter.AcquireConnection(clientName) {
			s.logger.Printf("Too many concurrent connections for client: %s", clientName)
			s.auditLogger.LogRequest(clientName, r.Method, r.URL.Path, false, http.StatusServiceUnavailable, time.Since(start))
			http.Error(w, "Too many concurrent connections", http.StatusServiceUnavailable)
			return
		}
		defer s.rateLimiter.ReleaseConnection(clientName)

		// Get or create client info
		client := s.getOrCreateClient(clientName)
		client.LastSeen = time.Now()
		client.RequestCount++

		s.logger.Printf("[%s] %s %s", clientName, r.Method, r.URL.Path)

		// Add client info to context
		ctx := context.WithValue(r.Context(), "client", client)
		ctx = context.WithValue(ctx, "client_name", clientName)
		
		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrapped, r.WithContext(ctx))
		
		// Log request to audit trail
		s.auditLogger.LogRequest(clientName, r.Method, r.URL.Path, wrapped.statusCode < 400, wrapped.statusCode, time.Since(start))
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// getClientName extracts the client name from the TLS certificate
func (s *Server) getClientName(r *http.Request) string {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return ""
	}

	cert := r.TLS.PeerCertificates[0]
	if len(cert.DNSNames) > 0 {
		return cert.DNSNames[0]
	}
	if cert.Subject.CommonName != "" {
		return cert.Subject.CommonName
	}
	return cert.SerialNumber.String()
}

// getOrCreateClient gets existing client info or creates new one
func (s *Server) getOrCreateClient(name string) *ClientInfo {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	if client, exists := s.clients[name]; exists {
		return client
	}

	client := &ClientInfo{
		Name:         name,
		LastSeen:     time.Now(),
		RateLimiter:  rate.NewLimiter(10, 20), // 10 req/s, burst 20
		AllowedScopes: []string{},
	}
	s.clients[name] = client
	return client
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"gt-proxy-server"}`))
}

// handleExec handles command execution requests
func (s *Server) handleExec(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	clientName := s.getClientName(r)
	
	if r.Method != http.MethodPost {
		s.auditLogger.LogExec(clientName, "", nil, false, http.StatusMethodNotAllowed, time.Since(start), nil)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req ExecRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		s.auditLogger.LogExec(clientName, "", nil, false, http.StatusBadRequest, time.Since(start), err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate and sanitize command/args
	if err := s.sanitizer.ValidateCommand(req.Command, req.Args); err != nil {
		s.logger.Printf("Security violation from %s: %v", clientName, err)
		s.auditLogger.LogExec(clientName, req.Command, req.Args, false, http.StatusBadRequest, time.Since(start), err)
		http.Error(w, fmt.Sprintf("Security violation: %v", err), http.StatusBadRequest)
		return
	}

	// Validate command against allowlist
	if s.allowlist != nil && !s.allowlist.IsAllowed(req.Command, req.Args) {
		s.logger.Printf("Blocked command from %s: %s %v", clientName, req.Command, req.Args)
		s.auditLogger.LogExec(clientName, req.Command, req.Args, false, http.StatusForbidden, time.Since(start), nil)
		http.Error(w, "Command not allowed", http.StatusForbidden)
		return
	}

	// Get command spec for validation
	spec, hasSpec := s.allowlist.GetSpec(req.Command)

	// Apply default timeout if not specified, or use spec timeout
	timeout := req.Timeout
	if timeout == 0 {
		timeout = s.sanitizer.RequestTimeout()
		if hasSpec && spec.Timeout > 0 {
			timeout = time.Duration(spec.Timeout) * time.Second
		}
	}

	// Execute command
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, req.Command, req.Args...)
	if req.WorkingDir != "" {
		// Validate working directory against allowed directories
		allowedDirs := []string{}
		if hasSpec {
			allowedDirs = spec.WorkingDirs
		}
		if err := security.SafePath(req.WorkingDir, allowedDirs); err != nil {
			s.auditLogger.LogExec(clientName, req.Command, req.Args, false, http.StatusBadRequest, time.Since(start), err)
			http.Error(w, fmt.Sprintf("Invalid working directory: %v", err), http.StatusBadRequest)
			return
		}
		cmd.Dir = req.WorkingDir
	}

	// Build environment: start with minimal safe defaults, add allowed vars
	cmd.Env = []string{
		"PATH=/usr/local/bin:/usr/bin:/bin",
		"HOME=/tmp",
		"USER=nobody",
	}
	for k, v := range req.Env {
		// Validate environment variable against allowlist
		if hasSpec {
			if !security.IsAllowedEnvVar(k, spec.EnvVars) {
				s.logger.Printf("Blocked env var from %s: %s", clientName, k)
				continue
			}
		} else {
			// No spec = block all env vars for safety
			s.logger.Printf("Blocked env var from %s (no spec): %s", clientName, k)
			continue
		}
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Capture output with size limit
	maxOutput := int64(10 * 1024 * 1024) // 10MB default
	if hasSpec && spec.MaxOutput > 0 {
		maxOutput = spec.MaxOutput
	}

	// Use LimitedReader to enforce output size cap
	output, err := s.execWithOutputLimit(ctx, cmd, maxOutput)

	resp := ExecResponse{
		ExitCode: 0,
		Output:   string(output),
	}
	if cmd.ProcessState != nil {
		resp.ExitCode = cmd.ProcessState.ExitCode()
	}
	if err != nil && ctx.Err() == nil {
		resp.Error = err.Error()
	}

	// Log successful execution to audit trail
	s.auditLogger.LogExec(clientName, req.Command, req.Args, true, http.StatusOK, time.Since(start), nil)

	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, resp)
}

// handleGit handles Git push/pull requests through the proxy
func (s *Server) handleGit(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	client := r.Context().Value("client").(*ClientInfo)
	clientName := s.getClientName(r)

	// Extract git operation from path
	// /git/push/<repo> or /git/pull/<repo>
	parts := strings.SplitN(r.URL.Path, "/", 4)
	if len(parts) < 3 {
		s.auditLogger.LogGit(clientName, "unknown", "", false, http.StatusBadRequest, time.Since(start), nil)
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	operation := parts[2] // push or pull
	repo := ""
	if len(parts) > 3 {
		repo = parts[3]
	}

	// Check if client is allowed to perform this operation on this repo
	if !s.isGitOperationAllowed(client, operation, repo) {
		s.auditLogger.LogGit(clientName, operation, repo, false, http.StatusForbidden, time.Since(start), nil)
		http.Error(w, "Git operation not allowed for this client", http.StatusForbidden)
		return
	}

	// Proxy the git operation
	// This would typically forward to the actual git server
	s.logger.Printf("Git %s for %s on repo %s", operation, client.Name, repo)

	// Log successful git operation to audit trail
	s.auditLogger.LogGit(clientName, operation, repo, true, http.StatusOK, time.Since(start), nil)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"status":"ok","operation":"%s","repo":"%s"}`, operation, repo)))
}

// handleGT handles gt command proxying
func (s *Server) handleGT(w http.ResponseWriter, r *http.Request) {
	// Proxy gt commands (bd, etc.)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok","service":"gt-proxy"}`))
}

// handleClientRegister handles client certificate registration
func (s *Server) handleClientRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate and sign client certificate
	clientCert, err := s.generateClientCert(req.ClientName, req.Scopes)
	if err != nil {
		s.logger.Printf("Failed to generate client cert: %v", err)
		http.Error(w, "Failed to generate certificate", http.StatusInternalServerError)
		return
	}

	resp := RegisterResponse{
		ClientName:  req.ClientName,
		Certificate: clientCert,
		ExpiresAt:   time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, resp)
}

// isGitOperationAllowed checks if a client is allowed to perform a git operation
func (s *Server) isGitOperationAllowed(client *ClientInfo, operation, repo string) bool {
	// Check branch scoping for push operations
	if operation == "push" {
		// Parse branch from request (would be in request body)
		// For now, check if client has any allowed scopes
		return len(client.AllowedScopes) > 0
	}
	return true
}

// generateClientCert generates a client certificate signed by the CA
func (s *Server) generateClientCert(clientName string, scopes []string) (string, error) {
	// This would generate a real certificate signed by the CA
	// For now, return a placeholder
	return "CLIENT_CERT_PLACEHOLDER", nil
}

// ExecRequest represents a command execution request
type ExecRequest struct {
	Command    string            `json:"command"`
	Args       []string          `json:"args"`
	WorkingDir string            `json:"working_dir"`
	Env        map[string]string `json:"env"`
	Timeout    time.Duration     `json:"timeout"`
}

// ExecResponse represents a command execution response
type ExecResponse struct {
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
}

// RegisterRequest represents a client registration request
type RegisterRequest struct {
	ClientName string   `json:"client_name"`
	Scopes     []string `json:"scopes"`
}

// RegisterResponse represents a client registration response
type RegisterResponse struct {
	ClientName  string `json:"client_name"`
	Certificate string `json:"certificate"`
	ExpiresAt   string `json:"expires_at"`
}

// decodeJSON decodes JSON from reader
func decodeJSON(r io.Reader, v interface{}) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

// encodeJSON encodes JSON to writer
func encodeJSON(w io.Writer, v interface{}) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(v)
}

// execWithOutputLimit executes a command and captures output with a size limit
func (s *Server) execWithOutputLimit(ctx context.Context, cmd *exec.Cmd, maxBytes int64) ([]byte, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// Read stdout with limit
	limitedStdout := io.LimitReader(stdout, maxBytes)
	stdoutData, _ := io.ReadAll(limitedStdout)

	// Read stderr with limit (remaining allowance)
	stderrLimit := maxBytes - int64(len(stdoutData))
	if stderrLimit < 0 {
		stderrLimit = 0
	}
	limitedStderr := io.LimitReader(stderr, stderrLimit)
	stderrData, _ := io.ReadAll(limitedStderr)

	// Wait for command to finish
	err = cmd.Wait()

	// Combine output
	output := stdoutData
	if len(stderrData) > 0 {
		if len(output) > 0 {
			output = append(output, '\n')
		}
		output = append(output, stderrData...)
	}

	return output, err
}
