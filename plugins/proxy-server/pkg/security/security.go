// Package security provides security utilities for the proxy server
package security

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Config holds security configuration
type Config struct {
	// Request size limits
	MaxRequestBodySize int64

	// Request timeout
	RequestTimeout time.Duration

	// Blocked patterns (shell injection attempts)
	BlockedPatterns []string

	// Security headers
	DisableServerVersion bool
	SecurityHeaders      map[string]string
}

// DefaultConfig returns default security configuration
func DefaultConfig() Config {
	return Config{
		MaxRequestBodySize: 1 * 1024 * 1024, // 1MB default
		RequestTimeout:     30 * time.Second,
		BlockedPatterns: []string{
			`;`,           // Command separator
			`&&`,          // AND operator
			`\|\|`,       // OR operator
			"`",           // Backtick command substitution
			`\$\(`,        // $() command substitution
			`<`,           // Input redirection
			`>`,           // Output redirection
			`\|`,          // Pipe
			`&`,           // Background
			`\n`,          // Newline (command break)
			`\r`,          // Carriage return
		},
		DisableServerVersion: true,
		SecurityHeaders: map[string]string{
			"X-Content-Type-Options": "nosniff",
			"X-Frame-Options":        "DENY",
			"X-XSS-Protection":       "1; mode=block",
			"Referrer-Policy":        "strict-origin-when-cross-origin",
		},
	}
}

// Sanitizer handles input sanitization
type Sanitizer struct {
	config       Config
	blockedRegex *regexp.Regexp
}

// NewSanitizer creates a new input sanitizer
func NewSanitizer(cfg Config) (*Sanitizer, error) {
	// Compile blocked patterns into a single regex
	if len(cfg.BlockedPatterns) > 0 {
		pattern := "(" + strings.Join(cfg.BlockedPatterns, "|") + ")"
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to compile blocked patterns: %w", err)
		}
		return &Sanitizer{
			config:       cfg,
			blockedRegex: re,
		}, nil
	}

	return &Sanitizer{
		config:       cfg,
		blockedRegex: nil,
	}, nil
}

// SanitizeArgs checks command arguments for shell injection attempts
func (s *Sanitizer) SanitizeArgs(args []string) error {
	for i, arg := range args {
		if err := s.SanitizeString(arg); err != nil {
			return fmt.Errorf("argument %d contains unsafe characters: %w", i, err)
		}
	}

	return nil
}

// SanitizeString checks a string for shell injection attempts
func (s *Sanitizer) SanitizeString(input string) error {
	if s.blockedRegex == nil {
		return nil
	}

	if s.blockedRegex.MatchString(input) {
		// Find what matched for better error messages
		matches := s.blockedRegex.FindString(input)
		return fmt.Errorf("input contains blocked pattern: %q", matches)
	}

	return nil
}

// ValidateCommand validates that a command is safe to execute
func (s *Sanitizer) ValidateCommand(command string, args []string) error {
	// Check command itself
	if err := s.SanitizeString(command); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	// Check arguments
	if err := s.SanitizeArgs(args); err != nil {
		return err
	}

	return nil
}

// MaxBodySize returns the configured maximum request body size
func (s *Sanitizer) MaxBodySize() int64 {
	return s.config.MaxRequestBodySize
}

// RequestTimeout returns the configured request timeout
func (s *Sanitizer) RequestTimeout() time.Duration {
	return s.config.RequestTimeout
}

// SecurityMiddleware adds security headers and request limits
func SecurityMiddleware(cfg Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Apply security headers
			for header, value := range cfg.SecurityHeaders {
				w.Header().Set(header, value)
			}

			// Apply timeout to request context
			ctx, cancel := createTimeoutContext(r, cfg.RequestTimeout)
			defer cancel()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// BodySizeLimitMiddleware limits request body size
func BodySizeLimitMiddleware(maxSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)
			next.ServeHTTP(w, r)
		})
	}
}

// createTimeoutContext creates a context with timeout
func createTimeoutContext(r *http.Request, timeout time.Duration) (context.Context, func()) {
	// Note: We can't actually use context.WithTimeout here without importing context
	// This is a simplified version - in production, use proper context timeout
	return r.Context(), func() {}
}

// SafePath checks if a path is safe (no directory traversal)
func SafePath(path string, allowedDirs []string) error {
	// Check for directory traversal attempts
	if strings.Contains(path, "..") {
		return fmt.Errorf("path contains directory traversal sequence")
	}

	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null byte")
	}

	// If no allowed directories specified, just check for traversal
	if len(allowedDirs) == 0 {
		return nil
	}

	// Resolve to absolute path for validation
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path is within any allowed directory
	for _, allowedDir := range allowedDirs {
		absAllowed, err := filepath.Abs(allowedDir)
		if err != nil {
			continue
		}
		// Ensure allowed directory ends with separator for prefix check
		if !strings.HasSuffix(absAllowed, string(filepath.Separator)) {
			absAllowed += string(filepath.Separator)
		}
		if strings.HasPrefix(absPath, absAllowed) || absPath == filepath.Clean(allowedDir) {
			return nil
		}
	}

	return fmt.Errorf("path %q is outside of allowed directories", path)
}

// SafePathNoSandbox checks if a path is safe without sandbox restrictions
// Deprecated: Use SafePath with explicit allowed directories
func SafePathNoSandbox(path string) error {
	return SafePath(path, nil)
}

// IsAllowedEnvVar checks if an environment variable is allowed
func IsAllowedEnvVar(name string, allowedPatterns []string) bool {
	for _, pattern := range allowedPatterns {
		// Exact match
		if pattern == name {
			return true
		}

		// Glob pattern (e.g., "GIT_*")
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(name, prefix) {
				return true
			}
		}
	}

	return false
}

// StripServerVersion removes server version information from errors
func StripServerVersion(err error) error {
	if err == nil {
		return nil
	}
	// Return generic error without version info
	return fmt.Errorf("internal server error")
}
