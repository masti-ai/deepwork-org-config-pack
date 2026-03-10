package ratelimit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Default.Default.RequestsPerSecond != 10 {
		t.Errorf("Expected default RPS 10, got %f", cfg.Default.Default.RequestsPerSecond)
	}

	if cfg.Default.Default.BurstSize != 20 {
		t.Errorf("Expected default burst 20, got %d", cfg.Default.Default.BurstSize)
	}

	if cfg.MaxConcurrentConnections != 10 {
		t.Errorf("Expected max connections 10, got %d", cfg.MaxConcurrentConnections)
	}

	// Check endpoint-specific limits
	execLimits, ok := cfg.Default.ByEndpoint[EndpointExec]
	if !ok {
		t.Error("Expected exec endpoint limits")
	} else {
		if execLimits.RequestsPerSecond != 1.0 { // 60/min = 1/sec
			t.Errorf("Expected exec RPS 1.0, got %f", execLimits.RequestsPerSecond)
		}
	}
}

func TestNew(t *testing.T) {
	cfg := DefaultConfig()
	limiter := New(cfg)

	if limiter == nil {
		t.Fatal("Expected non-nil limiter")
	}

	if limiter.clients == nil {
		t.Error("Expected clients map to be initialized")
	}
}

func TestAllow(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Default.Default.RequestsPerSecond = 1000 // Very high limit for testing
	cfg.Default.Default.BurstSize = 1000         // Large burst to ensure all requests pass
	// Clear endpoint-specific limits so they don't override defaults
	cfg.Default.ByEndpoint = map[EndpointType]Limits{}

	limiter := New(cfg)

	// Should allow first request
	if !limiter.Allow("client1", EndpointAPI) {
		t.Error("Expected first request to be allowed")
	}

	// Should allow subsequent requests (11 total)
	for i := 0; i < 10; i++ {
		if !limiter.Allow("client1", EndpointAPI) {
			t.Errorf("Expected request %d to be allowed", i+2)
		}
	}
}

func TestAllowPerClient(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Default.Default.RequestsPerSecond = 100
	cfg.Default.Default.BurstSize = 100

	limiter := New(cfg)

	// Client 1 uses up some requests
	for i := 0; i < 50; i++ {
		limiter.Allow("client1", EndpointAPI)
	}

	// Client 2 should still have full quota
	if !limiter.Allow("client2", EndpointAPI) {
		t.Error("Expected client2 request to be allowed (independent of client1)")
	}
}

func TestAllowPerEndpoint(t *testing.T) {
	cfg := DefaultConfig()
	// Set different limits for different endpoints
	cfg.Default.ByEndpoint[EndpointExec] = Limits{
		RequestsPerSecond: 0.1, // Very low limit
		BurstSize:         1,
	}
	cfg.Default.ByEndpoint[EndpointGit] = Limits{
		RequestsPerSecond: 100, // High limit
		BurstSize:         100,
	}

	limiter := New(cfg)

	// Use up exec limit
	if !limiter.Allow("client1", EndpointExec) {
		t.Error("Expected first exec request to be allowed")
	}

	// Second exec should be blocked (low rate)
	time.Sleep(10 * time.Millisecond)
	if limiter.Allow("client1", EndpointExec) {
		t.Error("Expected second exec request to be blocked (rate limit)")
	}

	// But git should still be allowed
	if !limiter.Allow("client1", EndpointGit) {
		t.Error("Expected git request to be allowed (different endpoint)")
	}
}

func TestAcquireConnection(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxConcurrentConnections = 2

	limiter := New(cfg)

	// Should acquire first connection
	if !limiter.AcquireConnection("client1") {
		t.Error("Expected first connection to be acquired")
	}

	// Should acquire second connection
	if !limiter.AcquireConnection("client1") {
		t.Error("Expected second connection to be acquired")
	}

	// Third should be blocked
	if limiter.AcquireConnection("client1") {
		t.Error("Expected third connection to be blocked")
	}
}

func TestReleaseConnection(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxConcurrentConnections = 1

	limiter := New(cfg)

	// Acquire and release
	if !limiter.AcquireConnection("client1") {
		t.Fatal("Expected first connection to be acquired")
	}

	limiter.ReleaseConnection("client1")

	// Should be able to acquire again
	if !limiter.AcquireConnection("client1") {
		t.Error("Expected connection to be acquired after release")
	}
}

func TestGetClientStats(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxConcurrentConnections = 5

	limiter := New(cfg)

	// Get stats for non-existent client
	_, _, ok := limiter.GetClientStats("nonexistent")
	if ok {
		t.Error("Expected GetClientStats to return false for non-existent client")
	}

	// Acquire connections and check stats
	limiter.AcquireConnection("client1")
	limiter.AcquireConnection("client1")

	activeConns, _, ok := limiter.GetClientStats("client1")
	if !ok {
		t.Error("Expected GetClientStats to return true for existing client")
	}

	if activeConns != 2 {
		t.Errorf("Expected 2 active connections, got %d", activeConns)
	}
}

func TestCleanup(t *testing.T) {
	cfg := DefaultConfig()
	limiter := New(cfg)

	// Create a client
	limiter.Allow("client1", EndpointAPI)

	// Verify client exists
	limiter.mu.RLock()
	_, exists := limiter.clients["client1"]
	limiter.mu.RUnlock()

	if !exists {
		t.Fatal("Expected client1 to exist")
	}

	// Wait a bit to ensure time passes
	time.Sleep(10 * time.Millisecond)

	// Cleanup with short max age - should remove stale client
	limiter.Cleanup(1 * time.Millisecond)

	// Client should be removed (lastSeen is older than maxAge and no active connections)
	limiter.mu.RLock()
	_, exists = limiter.clients["client1"]
	limiter.mu.RUnlock()

	if exists {
		t.Error("Client should have been cleaned up (stale with no active connections)")
	}

	// Test that client with active connections is not removed
	limiter2 := New(cfg)
	limiter2.Allow("client2", EndpointAPI)
	limiter2.AcquireConnection("client2")

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Try cleanup - should not remove client with active connections
	limiter2.Cleanup(1 * time.Millisecond)

	limiter2.mu.RLock()
	_, exists = limiter2.clients["client2"]
	limiter2.mu.RUnlock()

	if !exists {
		t.Error("Client with active connections should not be cleaned up")
	}

	// Release connection and cleanup again
	limiter2.ReleaseConnection("client2")
	limiter2.Cleanup(1 * time.Millisecond)

	limiter2.mu.RLock()
	_, exists = limiter2.clients["client2"]
	limiter2.mu.RUnlock()

	if exists {
		t.Error("Client should have been cleaned up after releasing connections")
	}
}

func TestSetClientLimits(t *testing.T) {
	cfg := DefaultConfig()
	limiter := New(cfg)

	// Create a client first
	limiter.Allow("client1", EndpointAPI)

	// Set custom limits
	newLimits := ClientLimits{
		Default: Limits{
			RequestsPerSecond: 50,
			BurstSize:         100,
		},
	}

	limiter.SetClientLimits("client1", newLimits)

	// Client should be removed and will be recreated with new limits
	limiter.mu.RLock()
	_, exists := limiter.clients["client1"]
	limiter.mu.RUnlock()

	if exists {
		t.Error("Expected client to be removed after setting new limits")
	}
}

func TestClassifyEndpoint(t *testing.T) {
	tests := []struct {
		path     string
		expected EndpointType
	}{
		{"/exec", EndpointExec},
		{"/exec/run", EndpointExec},
		{"/git/push", EndpointGit},
		{"/git/pull/repo", EndpointGit},
		{"/gt/status", EndpointGT},
		{"/health", EndpointAPI},
		{"/", EndpointAPI},
		{"", EndpointAPI},
	}

	for _, test := range tests {
		result := classifyEndpoint(test.path)
		if result != test.expected {
			t.Errorf("classifyEndpoint(%q) = %v, expected %v", test.path, result, test.expected)
		}
	}
}

func TestFormatRate(t *testing.T) {
	tests := []struct {
		rps      float64
		expected string
	}{
		{10, "10.0/sec"},
		{1, "1.0/sec"},
		{0.5, "30.0/min"},
		{0.1, "6.0/min"},
	}

	for _, test := range tests {
		result := FormatRate(test.rps)
		if result != test.expected {
			t.Errorf("FormatRate(%f) = %q, expected %q", test.rps, result, test.expected)
		}
	}
}

func TestMiddleware(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Default.Default.RequestsPerSecond = 100
	cfg.Default.Default.BurstSize = 100
	cfg.MaxConcurrentConnections = 10

	limiter := New(cfg)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with middleware
	middleware := limiter.Middleware(testHandler)

	// Test without client context - should fail
	req1 := httptest.NewRequest("GET", "/test", nil)
	rr1 := httptest.NewRecorder()
	middleware.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 without client, got %d", rr1.Code)
	}

	// Test with client context - should succeed
	req2 := httptest.NewRequest("GET", "/test", nil)
	ctx := req2.Context()
	ctx = context.WithValue(ctx, "client_name", "test-client")
	req2 = req2.WithContext(ctx)

	rr2 := httptest.NewRecorder()
	middleware.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("Expected 200 with client, got %d", rr2.Code)
	}
}


