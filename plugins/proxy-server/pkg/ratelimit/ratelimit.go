// Package ratelimit provides configurable rate limiting for the proxy server
package ratelimit

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// EndpointType represents the type of endpoint for rate limiting
type EndpointType string

const (
	EndpointExec EndpointType = "exec"
	EndpointGit  EndpointType = "git"
	EndpointGT   EndpointType = "gt"
	EndpointAPI  EndpointType = "api"
)

// Limits defines rate limits for a specific endpoint type
type Limits struct {
	RequestsPerSecond float64
	BurstSize         int
}

// ClientLimits stores per-client rate limit configuration
type ClientLimits struct {
	Default Limits
	ByEndpoint map[EndpointType]Limits
}

// Config holds rate limiter configuration
type Config struct {
	// Default limits apply to all clients unless overridden
	Default ClientLimits
	
	// Per-client overrides
	ClientOverrides map[string]ClientLimits
	
	// Max concurrent connections per client
	MaxConcurrentConnections int
}

// DefaultConfig returns a default rate limiting configuration
func DefaultConfig() Config {
	return Config{
		Default: ClientLimits{
			Default: Limits{
				RequestsPerSecond: 10,
				BurstSize:         20,
			},
			ByEndpoint: map[EndpointType]Limits{
				EndpointExec: {
					RequestsPerSecond: 60.0 / 60.0, // 60 per minute = 1 per second
					BurstSize:         5,
				},
				EndpointGit: {
					RequestsPerSecond: 30.0 / 60.0, // 30 per minute = 0.5 per second
					BurstSize:         3,
				},
				EndpointGT: {
					RequestsPerSecond: 20.0 / 60.0, // 20 per minute
					BurstSize:         5,
				},
				EndpointAPI: {
					RequestsPerSecond: 120.0 / 60.0, // 120 per minute = 2 per second
					BurstSize:         10,
				},
			},
		},
		ClientOverrides:          make(map[string]ClientLimits),
		MaxConcurrentConnections: 10,
	}
}

// Limiter manages rate limiting for multiple clients
type Limiter struct {
	config    Config
	clients   map[string]*clientState
	mu        sync.RWMutex
}

// clientState holds rate limiters and connection count for a client
type clientState struct {
	name               string
	limiters           map[EndpointType]*rate.Limiter
	defaultLimiter     *rate.Limiter
	activeConnections  int
	lastSeen           time.Time
	mu                 sync.Mutex
}

// New creates a new rate limiter
func New(cfg Config) *Limiter {
	return &Limiter{
		config:  cfg,
		clients: make(map[string]*clientState),
	}
}

// Allow checks if a request is allowed for the given client and endpoint
func (l *Limiter) Allow(clientName string, endpoint EndpointType) bool {
	state := l.getOrCreateClient(clientName)
	
	state.mu.Lock()
	defer state.mu.Unlock()
	
	state.lastSeen = time.Now()
	
	// Try endpoint-specific limiter first
	if limiter, ok := state.limiters[endpoint]; ok {
		return limiter.Allow()
	}
	
	// Fall back to default limiter
	return state.defaultLimiter.Allow()
}

// AcquireConnection attempts to acquire a connection slot for the client
// Returns true if the connection is allowed, false if max connections exceeded
func (l *Limiter) AcquireConnection(clientName string) bool {
	state := l.getOrCreateClient(clientName)
	
	state.mu.Lock()
	defer state.mu.Unlock()
	
	if state.activeConnections >= l.config.MaxConcurrentConnections {
		return false
	}
	
	state.activeConnections++
	state.lastSeen = time.Now()
	return true
}

// ReleaseConnection releases a connection slot for the client
func (l *Limiter) ReleaseConnection(clientName string) {
	state := l.getOrCreateClient(clientName)
	
	state.mu.Lock()
	defer state.mu.Unlock()
	
	if state.activeConnections > 0 {
		state.activeConnections--
	}
}

// GetClientStats returns rate limiting statistics for a client
func (l *Limiter) GetClientStats(clientName string) (activeConns int, lastSeen time.Time, ok bool) {
	l.mu.RLock()
	state, exists := l.clients[clientName]
	l.mu.RUnlock()
	
	if !exists {
		return 0, time.Time{}, false
	}
	
	state.mu.Lock()
	defer state.mu.Unlock()
	
	return state.activeConnections, state.lastSeen, true
}

// getOrCreateClient gets or creates a client state
func (l *Limiter) getOrCreateClient(name string) *clientState {
	l.mu.RLock()
	if state, ok := l.clients[name]; ok {
		l.mu.RUnlock()
		return state
	}
	l.mu.RUnlock()
	
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Double-check after acquiring write lock
	if state, ok := l.clients[name]; ok {
		return state
	}
	
	// Get client limits (check for override first)
	clientLimits := l.config.Default
	if override, ok := l.config.ClientOverrides[name]; ok {
		clientLimits = override
	}
	
	// Create limiters
	limiters := make(map[EndpointType]*rate.Limiter)
	for endpoint, limits := range clientLimits.ByEndpoint {
		limiters[endpoint] = rate.NewLimiter(rate.Limit(limits.RequestsPerSecond), limits.BurstSize)
	}
	
	state := &clientState{
		name:           name,
		limiters:       limiters,
		defaultLimiter: rate.NewLimiter(rate.Limit(clientLimits.Default.RequestsPerSecond), clientLimits.Default.BurstSize),
		lastSeen:       time.Now(),
	}
	
	l.clients[name] = state
	return state
}

// Cleanup removes stale client entries
func (l *Limiter) Cleanup(maxAge time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	cutoff := time.Now().Add(-maxAge)
	for name, state := range l.clients {
		state.mu.Lock()
		if state.lastSeen.Before(cutoff) && state.activeConnections == 0 {
			delete(l.clients, name)
		}
		state.mu.Unlock()
	}
}

// SetClientLimits updates rate limits for a specific client
func (l *Limiter) SetClientLimits(clientName string, limits ClientLimits) {
	l.mu.Lock()
	l.config.ClientOverrides[clientName] = limits
	l.mu.Unlock()
	
	// Remove existing state to force recreation with new limits
	l.mu.Lock()
	delete(l.clients, clientName)
	l.mu.Unlock()
}

// Middleware returns HTTP middleware that applies rate limiting
func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract client name from context (set by auth middleware)
		clientName, ok := r.Context().Value("client_name").(string)
		if !ok || clientName == "" {
			http.Error(w, "Client identification required", http.StatusUnauthorized)
			return
		}
		
		// Determine endpoint type from path
		endpoint := classifyEndpoint(r.URL.Path)
		
		// Check rate limit
		if !l.Allow(clientName, endpoint) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		
		// Check concurrent connections
		if !l.AcquireConnection(clientName) {
			http.Error(w, "Too many concurrent connections", http.StatusServiceUnavailable)
			return
		}
		defer l.ReleaseConnection(clientName)
		
		next.ServeHTTP(w, r)
	})
}

// classifyEndpoint determines the endpoint type from the request path
func classifyEndpoint(path string) EndpointType {
	if len(path) < 1 {
		return EndpointAPI
	}
	
	if path[0] == '/' {
		path = path[1:]
	}
	
	switch {
	case len(path) >= 4 && path[:4] == "exec":
		return EndpointExec
	case len(path) >= 3 && path[:3] == "git":
		return EndpointGit
	case len(path) >= 2 && path[:2] == "gt":
		return EndpointGT
	default:
		return EndpointAPI
	}
}

// FormatRate formats a rate as "X/min" or "X/sec" for display
func FormatRate(rps float64) string {
	if rps >= 1 {
		return fmt.Sprintf("%.1f/sec", rps)
	}
	return fmt.Sprintf("%.1f/min", rps*60)
}
