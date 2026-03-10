// gt-proxy-server provides mTLS-secured command proxying for sandboxed containers
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Deepwork-AI/ai-planogram/proxy-server/pkg/allowlist"
	"github.com/Deepwork-AI/ai-planogram/proxy-server/pkg/audit"
	"github.com/Deepwork-AI/ai-planogram/proxy-server/pkg/cert"
	"github.com/Deepwork-AI/ai-planogram/proxy-server/pkg/ratelimit"
	"github.com/Deepwork-AI/ai-planogram/proxy-server/pkg/security"
	"github.com/Deepwork-AI/ai-planogram/proxy-server/pkg/server"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	addr       string
	caCertPath string
	caKeyPath  string
	certDir    string
	rateLimit  float64
	
	// Audit flags
	auditLogPath string
	
	// Security flags
	maxBodySize    int64
	requestTimeout int

	rootCmd = &cobra.Command{
		Use:   "gt-proxy-server",
		Short: "mTLS-secured command proxy for sandboxed containers",
		Long: `gt-proxy-server provides secure command execution proxying
for sandboxed containers with mTLS authentication, rate limiting,
and command allowlisting.`,
	}

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Start the proxy server",
		RunE:  runServer,
	}

	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize CA and server certificates",
		RunE:  initCA,
	}

	clientCmd = &cobra.Command{
		Use:   "client",
		Short: "Client certificate management",
	}

	clientGenCmd = &cobra.Command{
		Use:   "generate [client-name]",
		Short: "Generate a client certificate",
		Args:  cobra.ExactArgs(1),
		RunE:  generateClientCert,
	}

	clientListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all client certificates",
		RunE:  listClientCerts,
	}
)

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&certDir, "cert-dir", "", "Certificate directory (default: ~/gt/.runtime/ca)")

	// Run command flags
	runCmd.Flags().StringVar(&addr, "addr", ":9876", "Server listen address")
	runCmd.Flags().StringVar(&caCertPath, "ca-cert", "", "CA certificate path")
	runCmd.Flags().StringVar(&caKeyPath, "ca-key", "", "CA key path")
	runCmd.Flags().Float64Var(&rateLimit, "rate-limit", 10, "Default rate limit per client (requests per second)")
	runCmd.Flags().StringVar(&auditLogPath, "audit-log", "", "Audit log file path (default: ~/gt/.runtime/logs/audit.jsonl)")
	runCmd.Flags().Int64Var(&maxBodySize, "max-body-size", 1024*1024, "Maximum request body size in bytes")
	runCmd.Flags().IntVar(&requestTimeout, "request-timeout", 30, "Request timeout in seconds")

	// Add commands
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(clientCmd)
	clientCmd.AddCommand(clientGenCmd)
	clientCmd.AddCommand(clientListCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func getCertPaths() (certDir, caCertPath, caKeyPath string, err error) {
	if certDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get home directory: %w", err)
		}
		certDir = filepath.Join(home, "gt", ".runtime", "ca")
	}

	if caCertPath == "" {
		caCertPath = filepath.Join(certDir, "ca.crt")
	}
	if caKeyPath == "" {
		caKeyPath = filepath.Join(certDir, "ca.key")
	}

	return certDir, caCertPath, caKeyPath, nil
}

func runServer(cmd *cobra.Command, args []string) error {
	certDir, caCertPath, caKeyPath, err := getCertPaths()
	if err != nil {
		return err
	}

	// Check if CA certificates exist
	if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
		return fmt.Errorf("CA certificate not found at %s. Run 'gt-proxy-server init' first", caCertPath)
	}

	// Create allowlist
	allowlist := allowlist.New()

	// Configure rate limiting (per-client, per-endpoint)
	rateLimitConfig := ratelimit.DefaultConfig()
	// Update default rate limit from flag
	rateLimitConfig.Default.Default.RequestsPerSecond = rateLimit
	rateLimitConfig.Default.Default.BurstSize = int(rateLimit * 2)
	
	// Configure audit logging
	auditLog := auditLogPath
	if auditLog == "" {
		auditLog = filepath.Join(certDir, "..", "logs", "audit.jsonl")
	}
	auditConfig := audit.Config{
		LogPath: auditLog,
	}

	// Configure security
	securityConfig := security.DefaultConfig()
	securityConfig.MaxRequestBodySize = maxBodySize
	securityConfig.RequestTimeout = time.Duration(requestTimeout) * time.Second

	// Create server
	cfg := server.Config{
		Addr:           addr,
		CACertPath:     caCertPath,
		CAKeyPath:      caKeyPath,
		CertDir:        certDir,
		RateLimit:      rateLimitConfig,
		Allowlist:      allowlist,
		AuditConfig:    auditConfig,
		SecurityConfig: securityConfig,
	}

	srv, err := server.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	return srv.Start()
}

func initCA(cmd *cobra.Command, args []string) error {
	certDir, _, _, err := getCertPaths()
	if err != nil {
		return err
	}

	log.Printf("Initializing CA in %s", certDir)

	certPath, keyPath, err := cert.GenerateCA(certDir)
	if err != nil {
		return fmt.Errorf("failed to generate CA: %w", err)
	}

	log.Printf("CA certificate generated: %s", certPath)
	log.Printf("CA key generated: %s", keyPath)
	log.Println("CA initialized successfully!")

	return nil
}

func generateClientCert(cmd *cobra.Command, args []string) error {
	clientName := args[0]
	certDir, caCertPath, caKeyPath, err := getCertPaths()
	if err != nil {
		return err
	}

	// Check if CA exists
	if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
		return fmt.Errorf("CA certificate not found. Run 'gt-proxy-server init' first")
	}

	// Create cert manager
	mgr, err := cert.NewManager(cert.Config{
		CACertPath: caCertPath,
		CAKeyPath:  caKeyPath,
		CertDir:    certDir,
	})
	if err != nil {
		return fmt.Errorf("failed to create cert manager: %w", err)
	}

	// Generate certificate
	scopes := []string{"git:read", "git:write", "exec:allowed"}
	certPEM, keyPEM, err := mgr.GenerateClientCert(clientName, scopes)
	if err != nil {
		return fmt.Errorf("failed to generate client certificate: %w", err)
	}

	// Save certificate
	certPath, keyPath, err := mgr.SaveClientCert(clientName, certPEM, keyPEM)
	if err != nil {
		return fmt.Errorf("failed to save client certificate: %w", err)
	}

	log.Printf("Client certificate generated for: %s", clientName)
	log.Printf("  Certificate: %s", certPath)
	log.Printf("  Key: %s", keyPath)

	return nil
}

func listClientCerts(cmd *cobra.Command, args []string) error {
	certDir, caCertPath, caKeyPath, err := getCertPaths()
	if err != nil {
		return err
	}

	mgr, err := cert.NewManager(cert.Config{
		CACertPath: caCertPath,
		CAKeyPath:  caKeyPath,
		CertDir:    certDir,
	})
	if err != nil {
		return fmt.Errorf("failed to create cert manager: %w", err)
	}

	clients, err := mgr.ListClientCerts()
	if err != nil {
		return fmt.Errorf("failed to list clients: %w", err)
	}

	if len(clients) == 0 {
		log.Println("No client certificates found")
		return nil
	}

	log.Println("Client certificates:")
	for _, client := range clients {
		log.Printf("  - %s", client)
	}

	return nil
}
