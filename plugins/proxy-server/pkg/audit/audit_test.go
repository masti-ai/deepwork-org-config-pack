package audit

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	// Test with log file
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.jsonl")

	cfg := Config{
		LogPath: logPath,
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer logger.Close()

	// Verify log file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Audit log file was not created")
	}
}

func TestNewWithStdout(t *testing.T) {
	// Test with stdout (no log path)
	cfg := Config{}

	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer logger.Close()
}

func TestLog(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.jsonl")

	cfg := Config{
		LogPath: logPath,
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer logger.Close()

	entry := Entry{
		Timestamp: time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
		ClientCN:  "test-client",
		Endpoint:  "exec",
		Command:   "ls",
		Allowed:   true,
		StatusCode: 200,
		DurationMs: 100,
	}

	if err := logger.Log(entry); err != nil {
		t.Fatalf("Failed to log entry: %v", err)
	}

	// Read and verify log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	var loggedEntry Entry
	if err := json.Unmarshal(content, &loggedEntry); err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	if loggedEntry.ClientCN != "test-client" {
		t.Errorf("Expected client CN 'test-client', got '%s'", loggedEntry.ClientCN)
	}

	if loggedEntry.Command != "ls" {
		t.Errorf("Expected command 'ls', got '%s'", loggedEntry.Command)
	}
}

func TestLogRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		writer:  &buf,
		encoder: json.NewEncoder(&buf),
	}

	err := logger.LogRequest("test-client", "POST", "/exec", true, 200, 150*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to log request: %v", err)
	}

	var entry Entry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse entry: %v", err)
	}

	if entry.ClientCN != "test-client" {
		t.Errorf("Expected client CN 'test-client', got '%s'", entry.ClientCN)
	}

	if entry.DurationMs != 150 {
		t.Errorf("Expected duration 150ms, got %d", entry.DurationMs)
	}
}

func TestLogExec(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		writer:  &buf,
		encoder: json.NewEncoder(&buf),
	}

	err := logger.LogExec("test-client", "git", []string{"status"}, true, 200, 50*time.Millisecond, nil)
	if err != nil {
		t.Fatalf("Failed to log exec: %v", err)
	}

	var entry Entry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse entry: %v", err)
	}

	if entry.Command != "git" {
		t.Errorf("Expected command 'git', got '%s'", entry.Command)
	}

	if len(entry.Args) != 1 || entry.Args[0] != "status" {
		t.Errorf("Expected args ['status'], got %v", entry.Args)
	}
}

func TestLogExecDenied(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		writer:  &buf,
		encoder: json.NewEncoder(&buf),
	}

	err := logger.LogExec("test-client", "rm", []string{"-rf", "/"}, false, 403, 0, nil)
	if err != nil {
		t.Fatalf("Failed to log exec: %v", err)
	}

	var entry Entry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse entry: %v", err)
	}

	if entry.Allowed {
		t.Error("Expected entry to be not allowed")
	}

	if entry.DeniedReason != "command not in allowlist" {
		t.Errorf("Expected denied reason, got '%s'", entry.DeniedReason)
	}
}

func TestLogGit(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		writer:  &buf,
		encoder: json.NewEncoder(&buf),
	}

	err := logger.LogGit("test-client", "push", "my-repo", true, 200, 500*time.Millisecond, nil)
	if err != nil {
		t.Fatalf("Failed to log git: %v", err)
	}

	var entry Entry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse entry: %v", err)
	}

	if entry.Endpoint != "git" {
		t.Errorf("Expected endpoint 'git', got '%s'", entry.Endpoint)
	}

	if entry.Metadata == nil {
		t.Fatal("Expected metadata to be set")
	}

	if entry.Metadata["operation"] != "push" {
		t.Errorf("Expected operation 'push', got '%v'", entry.Metadata["operation"])
	}
}

func TestExtractEndpoint(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/health", "health"},
		{"/exec", "exec"},
		{"/git/push", "git"},
		{"/gt/status", "gt"},
		{"/", ""},
		{"", "unknown"},
		{"/very/long/path", "very"},
	}

	for _, test := range tests {
		result := extractEndpoint(test.path)
		if result != test.expected {
			t.Errorf("extractEndpoint(%q) = %q, expected %q", test.path, result, test.expected)
		}
	}
}

func TestSetWebhook(t *testing.T) {
	logger := &Logger{}
	
	logger.SetWebhook("https://example.com/webhook")
	
	logger.webhookMu.RLock()
	webhook := logger.webhook
	logger.webhookMu.RUnlock()
	
	if webhook != "https://example.com/webhook" {
		t.Errorf("Expected webhook URL, got '%s'", webhook)
	}
}

func TestClose(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.jsonl")

	cfg := Config{
		LogPath: logPath,
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}

	if err := logger.Close(); err != nil {
		t.Fatalf("Failed to close logger: %v", err)
	}

	// Test closing stdout logger (no file)
	logger2 := &Logger{
		writer: os.Stdout,
		file:   nil,
	}

	if err := logger2.Close(); err != nil {
		t.Fatalf("Failed to close logger without file: %v", err)
	}
}
