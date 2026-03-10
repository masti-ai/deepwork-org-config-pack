package security

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MaxRequestBodySize != 1024*1024 {
		t.Errorf("Expected max body size 1MB, got %d", cfg.MaxRequestBodySize)
	}

	if cfg.RequestTimeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", cfg.RequestTimeout)
	}

	if len(cfg.BlockedPatterns) == 0 {
		t.Error("Expected blocked patterns to be configured")
	}

	if !cfg.DisableServerVersion {
		t.Error("Expected server version to be disabled by default")
	}

	if len(cfg.SecurityHeaders) == 0 {
		t.Error("Expected security headers to be configured")
	}
}

func TestNewSanitizer(t *testing.T) {
	cfg := DefaultConfig()
	sanitizer, err := NewSanitizer(cfg)
	if err != nil {
		t.Fatalf("Failed to create sanitizer: %v", err)
	}

	if sanitizer == nil {
		t.Fatal("Expected non-nil sanitizer")
	}

	if sanitizer.blockedRegex == nil {
		t.Error("Expected blocked regex to be compiled")
	}
}

func TestNewSanitizerEmptyPatterns(t *testing.T) {
	cfg := Config{
		MaxRequestBodySize: 1024 * 1024,
		RequestTimeout:     30 * time.Second,
		BlockedPatterns:    []string{}, // Empty patterns
	}

	sanitizer, err := NewSanitizer(cfg)
	if err != nil {
		t.Fatalf("Failed to create sanitizer: %v", err)
	}

	if sanitizer.blockedRegex != nil {
		t.Error("Expected nil regex when no patterns configured")
	}
}

func TestNewSanitizerInvalidPattern(t *testing.T) {
	cfg := Config{
		BlockedPatterns: []string{"[invalid"}, // Invalid regex
	}

	_, err := NewSanitizer(cfg)
	if err == nil {
		t.Error("Expected error for invalid regex pattern")
	}
}

func TestSanitizeString(t *testing.T) {
	cfg := DefaultConfig()
	sanitizer, _ := NewSanitizer(cfg)

	tests := []struct {
		input    string
		expected bool // true = should pass, false = should fail
	}{
		{"hello world", true},
		{"git status", true},
		{"ls -la", true},
		{"echo test", true},
		{"cmd; rm -rf /", false},       // Command separator
		{"cmd && rm -rf /", false},      // AND operator
		{"cmd || rm -rf /", false},      // OR operator
		{"`whoami`", false},              // Backtick substitution
		{"$(whoami)", false},             // $() substitution
		{"cmd > /etc/passwd", false},     // Output redirection
		{"cmd < /etc/passwd", false},     // Input redirection
		{"cmd | cat", false},             // Pipe
		{"cmd &", false},                 // Background
		{"line1\nline2", false},          // Newline
		{"line1\rline2", false},          // Carriage return
	}

	for _, test := range tests {
		err := sanitizer.SanitizeString(test.input)
		if test.expected && err != nil {
			t.Errorf("SanitizeString(%q) failed unexpectedly: %v", test.input, err)
		}
		if !test.expected && err == nil {
			t.Errorf("SanitizeString(%q) should have failed but passed", test.input)
		}
	}
}

func TestSanitizeStringNoPatterns(t *testing.T) {
	cfg := Config{
		MaxRequestBodySize: 1024 * 1024,
		RequestTimeout:     30 * time.Second,
		BlockedPatterns:    []string{},
	}
	sanitizer, _ := NewSanitizer(cfg)

	// Should pass when no patterns configured
	err := sanitizer.SanitizeString("anything; here")
	if err != nil {
		t.Errorf("SanitizeString failed with no patterns: %v", err)
	}
}

func TestSanitizeArgs(t *testing.T) {
	cfg := DefaultConfig()
	sanitizer, _ := NewSanitizer(cfg)

	// Valid args
	err := sanitizer.SanitizeArgs([]string{"-la", "/home", "test.txt"})
	if err != nil {
		t.Errorf("SanitizeArgs failed for valid args: %v", err)
	}

	// Invalid arg (contains semicolon)
	err = sanitizer.SanitizeArgs([]string{"-la", "/home; rm -rf /"})
	if err == nil {
		t.Error("SanitizeArgs should have failed for args with shell injection")
	}
}

func TestSanitizeArgsEmpty(t *testing.T) {
	cfg := DefaultConfig()
	sanitizer, _ := NewSanitizer(cfg)

	err := sanitizer.SanitizeArgs([]string{})
	if err != nil {
		t.Errorf("SanitizeArgs failed for empty args: %v", err)
	}
}

func TestValidateCommand(t *testing.T) {
	cfg := DefaultConfig()
	sanitizer, _ := NewSanitizer(cfg)

	// Valid command
	err := sanitizer.ValidateCommand("git", []string{"status", "--short"})
	if err != nil {
		t.Errorf("ValidateCommand failed for valid command: %v", err)
	}

	// Invalid command (contains blocked char)
	err = sanitizer.ValidateCommand("git;rm", []string{"-rf", "/"})
	if err == nil {
		t.Error("ValidateCommand should have failed for command with injection")
	}

	// Invalid arg
	err = sanitizer.ValidateCommand("echo", []string{"hello; rm -rf /"})
	if err == nil {
		t.Error("ValidateCommand should have failed for args with injection")
	}
}

func TestMaxBodySize(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRequestBodySize = 5 * 1024 * 1024 // 5MB
	sanitizer, _ := NewSanitizer(cfg)

	if sanitizer.MaxBodySize() != 5*1024*1024 {
		t.Errorf("Expected max body size 5MB, got %d", sanitizer.MaxBodySize())
	}
}

func TestRequestTimeout(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RequestTimeout = 60 * time.Second
	sanitizer, _ := NewSanitizer(cfg)

	if sanitizer.RequestTimeout() != 60*time.Second {
		t.Errorf("Expected timeout 60s, got %v", sanitizer.RequestTimeout())
	}
}

func TestSecurityMiddleware(t *testing.T) {
	cfg := DefaultConfig()

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with security middleware
	middleware := SecurityMiddleware(cfg)(testHandler)

	// Test request
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	// Check security headers
	headers := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}

	for header, expected := range headers {
		value := rr.Header().Get(header)
		if value != expected {
			t.Errorf("Expected header %s = %q, got %q", header, expected, value)
		}
	}
}

func TestBodySizeLimitMiddleware(t *testing.T) {
	maxSize := int64(100)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, 200)
		n, err := r.Body.Read(body)
		if err != nil && err.Error() != "http: request body too large" {
			t.Errorf("Expected 'request body too large' error, got: %v", err)
		}
		if n > int(maxSize) {
			t.Error("Read more bytes than max size allowed")
		}
	})

	// Wrap with body size limit middleware
	middleware := BodySizeLimitMiddleware(maxSize)(testHandler)

	// Test with body larger than limit
	req := httptest.NewRequest("POST", "/test", strings.NewReader(strings.Repeat("x", 200)))
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)
}

func TestSafePath(t *testing.T) {
	tests := []struct {
		path       string
		allowedDirs []string
		expected   bool // true = should pass, false = should fail
	}{
		{"/home/user/project", nil, true},
		{"/workspace/gt/test", nil, true},
		{"./relative/path", nil, true},
		{"../parent", nil, false},                              // Directory traversal
		{"/etc/../etc/passwd", nil, false},                     // Directory traversal
		{"/path/with/\x00/null", nil, false},                   // Null byte
		{"normal/path", nil, true},
		{"/workspace/gt/test", []string{"/workspace"}, true},   // In allowed dir
		{"/etc/passwd", []string{"/workspace"}, false},         // Outside allowed dir
		{"/tmp/file", []string{"/workspace", "/tmp"}, true},    // In second allowed dir
	}

	for _, test := range tests {
		err := SafePath(test.path, test.allowedDirs)
		if test.expected && err != nil {
			t.Errorf("SafePath(%q, %v) failed unexpectedly: %v", test.path, test.allowedDirs, err)
		}
		if !test.expected && err == nil {
			t.Errorf("SafePath(%q, %v) should have failed but passed", test.path, test.allowedDirs)
		}
	}
}

func TestIsAllowedEnvVar(t *testing.T) {
	allowedPatterns := []string{
		"HOME",
		"PATH",
		"GIT_*",
		"USER",
	}

	tests := []struct {
		name     string
		expected bool
	}{
		{"HOME", true},
		{"PATH", true},
		{"USER", true},
		{"GIT_CONFIG", true},
		{"GIT_TRACE", true},
		{"HTTP_PROXY", false},
		{"LD_PRELOAD", false},
	}

	for _, test := range tests {
		result := IsAllowedEnvVar(test.name, allowedPatterns)
		if result != test.expected {
			t.Errorf("IsAllowedEnvVar(%q) = %v, expected %v", test.name, result, test.expected)
		}
	}
}

func TestIsAllowedEnvVarEmptyPatterns(t *testing.T) {
	// With empty patterns, nothing should be allowed
	result := IsAllowedEnvVar("HOME", []string{})
	if result {
		t.Error("IsAllowedEnvVar should return false for empty patterns")
	}
}

func TestStripServerVersion(t *testing.T) {
	// Test with nil error
	result := StripServerVersion(nil)
	if result != nil {
		t.Errorf("StripServerVersion(nil) should return nil, got %v", result)
	}

	// Test with error - should return generic error
	originalErr := http.ErrBodyNotAllowed
	strippedErr := StripServerVersion(originalErr)
	if strippedErr == nil {
		t.Fatal("StripServerVersion should return non-nil for non-nil error")
	}
	if strippedErr.Error() != "internal server error" {
		t.Errorf("Expected 'internal server error', got %q", strippedErr.Error())
	}
}
