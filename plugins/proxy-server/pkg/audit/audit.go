// Package audit provides request logging and audit trail for the proxy server
package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Entry represents a single audit log entry
type Entry struct {
	Timestamp    time.Time              `json:"timestamp"`
	RequestID    string                 `json:"request_id"`
	ClientCN     string                 `json:"client_cn"`
	Endpoint     string                 `json:"endpoint"`
	Method       string                 `json:"method"`
	Path         string                 `json:"path"`
	Command      string                 `json:"command,omitempty"`
	Args         []string               `json:"args,omitempty"`
	Allowed      bool                   `json:"allowed"`
	DeniedReason string                 `json:"denied_reason,omitempty"`
	DurationMs   int64                  `json:"duration_ms"`
	StatusCode   int                    `json:"status_code"`
	Error        string                 `json:"error,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Logger handles audit log writing
type Logger struct {
	writer    io.Writer
	file      *os.File
	encoder   *json.Encoder
	mu        sync.Mutex
	webhook   string
	webhookMu sync.RWMutex
}

// Config holds audit logger configuration
type Config struct {
	LogPath    string
	WebhookURL string
	BufferSize int
}

// New creates a new audit logger
func New(cfg Config) (*Logger, error) {
	var writer io.Writer
	var file *os.File

	if cfg.LogPath != "" {
		// Ensure directory exists
		dir := filepath.Dir(cfg.LogPath)
		if err := os.MkdirAll(dir, 0750); err != nil {
			return nil, fmt.Errorf("failed to create audit log directory: %w", err)
		}

		// Open log file for appending
		f, err := os.OpenFile(cfg.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			return nil, fmt.Errorf("failed to open audit log file: %w", err)
		}
		file = f
		writer = f
	} else {
		// Default to stdout if no log path specified
		writer = os.Stdout
	}

	return &Logger{
		writer:  writer,
		file:    file,
		encoder: json.NewEncoder(writer),
		webhook: cfg.WebhookURL,
	}, nil
}

// Log writes an audit entry
func (l *Logger) Log(entry Entry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	// Write to file/stdout
	if err := l.encoder.Encode(entry); err != nil {
		return fmt.Errorf("failed to encode audit entry: %w", err)
	}

	// Optionally send to webhook (async)
	l.webhookMu.RLock()
	webhook := l.webhook
	l.webhookMu.RUnlock()

	if webhook != "" {
		go l.sendWebhook(entry)
	}

	return nil
}

// LogRequest is a helper to log an HTTP request
func (l *Logger) LogRequest(clientCN, method, path string, allowed bool, statusCode int, duration time.Duration) error {
	return l.Log(Entry{
		Timestamp:  time.Now().UTC(),
		ClientCN:   clientCN,
		Method:     method,
		Path:       path,
		Endpoint:   extractEndpoint(path),
		Allowed:    allowed,
		StatusCode: statusCode,
		DurationMs: duration.Milliseconds(),
	})
}

// LogExec is a helper to log a command execution
func (l *Logger) LogExec(clientCN, command string, args []string, allowed bool, statusCode int, duration time.Duration, err error) error {
	entry := Entry{
		Timestamp:  time.Now().UTC(),
		ClientCN:   clientCN,
		Endpoint:   "exec",
		Command:    command,
		Args:       args,
		Allowed:    allowed,
		StatusCode: statusCode,
		DurationMs: duration.Milliseconds(),
	}

	if !allowed {
		entry.DeniedReason = "command not in allowlist"
	}
	if err != nil {
		entry.Error = err.Error()
	}

	return l.Log(entry)
}

// LogGit is a helper to log a git operation
func (l *Logger) LogGit(clientCN, operation, repo string, allowed bool, statusCode int, duration time.Duration, err error) error {
	entry := Entry{
		Timestamp:  time.Now().UTC(),
		ClientCN:   clientCN,
		Endpoint:   "git",
		Command:    fmt.Sprintf("git %s", operation),
		Allowed:    allowed,
		StatusCode: statusCode,
		DurationMs: duration.Milliseconds(),
		Metadata: map[string]interface{}{
			"operation": operation,
			"repo":      repo,
		},
	}

	if !allowed {
		entry.DeniedReason = "git operation not allowed for client"
	}
	if err != nil {
		entry.Error = err.Error()
	}

	return l.Log(entry)
}

// SetWebhook updates the webhook URL
func (l *Logger) SetWebhook(url string) {
	l.webhookMu.Lock()
	defer l.webhookMu.Unlock()
	l.webhook = url
}

// Close closes the audit logger
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// sendWebhook sends the entry to the configured webhook (async)
func (l *Logger) sendWebhook(entry Entry) {
	// Implementation for webhook sending would go here
	// For now, this is a placeholder for future webhook integration
	// In production, this would make an HTTP POST to l.webhook
}

// extractEndpoint extracts the endpoint type from the path
func extractEndpoint(path string) string {
	if len(path) < 1 {
		return "unknown"
	}

	// Remove leading slash
	if path[0] == '/' {
		path = path[1:]
	}

	// Extract first segment
	for i, c := range path {
		if c == '/' {
			return path[:i]
		}
	}

	return path
}
