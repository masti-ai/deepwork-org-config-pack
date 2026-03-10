// Package cert handles certificate generation and management for mTLS
package cert

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Manager handles certificate generation and storage
type Manager struct {
	caCertPath string
	caKeyPath  string
	certDir    string
}

// Config holds certificate manager configuration
type Config struct {
	CACertPath string
	CAKeyPath  string
	CertDir    string
}

// NewManager creates a new certificate manager
func NewManager(cfg Config) (*Manager, error) {
	if cfg.CACertPath == "" || cfg.CAKeyPath == "" {
		return nil, fmt.Errorf("CA certificate and key paths are required")
	}
	if cfg.CertDir == "" {
		cfg.CertDir = filepath.Join(os.Getenv("HOME"), "gt", ".runtime", "certs")
	}

	// Ensure cert directory exists
	if err := os.MkdirAll(cfg.CertDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create cert directory: %w", err)
	}

	return &Manager{
		caCertPath: cfg.CACertPath,
		caKeyPath:  cfg.CAKeyPath,
		certDir:    cfg.CertDir,
	}, nil
}

// GenerateCA generates a new Certificate Authority
func GenerateCA(certDir string) (certPath, keyPath string, err error) {
	if err := os.MkdirAll(certDir, 0700); err != nil {
		return "", "", fmt.Errorf("failed to create CA directory: %w", err)
	}

	// Generate RSA key for CA
	caKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate CA key: %w", err)
	}

	// Create CA certificate template
	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"GT Proxy"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    "GT Proxy CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Self-sign the CA certificate
	caCertBytes, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to create CA certificate: %w", err)
	}

	// Write CA certificate
	certPath = filepath.Join(certDir, "ca.crt")
	certFile, err := os.Create(certPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create CA cert file: %w", err)
	}
	defer certFile.Close()

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: caCertBytes}); err != nil {
		return "", "", fmt.Errorf("failed to write CA certificate: %w", err)
	}

	// Write CA key
	keyPath = filepath.Join(certDir, "ca.key")
	keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", "", fmt.Errorf("failed to create CA key file: %w", err)
	}
	defer keyFile.Close()

	caKeyBytes, err := x509.MarshalPKCS8PrivateKey(caKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal CA key: %w", err)
	}

	if err := pem.Encode(keyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: caKeyBytes}); err != nil {
		return "", "", fmt.Errorf("failed to write CA key: %w", err)
	}

	return certPath, keyPath, nil
}

// GenerateClientCert generates a client certificate signed by the CA
func (m *Manager) GenerateClientCert(clientName string, scopes []string) (certPEM, keyPEM string, err error) {
	// Load CA certificate
	caCertPEM, err := os.ReadFile(m.caCertPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caBlock, _ := pem.Decode(caCertPEM)
	if caBlock == nil {
		return "", "", fmt.Errorf("failed to decode CA certificate")
	}

	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Load CA key
	caKeyPEM, err := os.ReadFile(m.caKeyPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read CA key: %w", err)
	}

	caKeyBlock, _ := pem.Decode(caKeyPEM)
	if caKeyBlock == nil {
		return "", "", fmt.Errorf("failed to decode CA key")
	}

	caKey, err := x509.ParsePKCS8PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse CA key: %w", err)
	}

	// Generate client key
	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate client key: %w", err)
	}

	// Create client certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return "", "", fmt.Errorf("failed to generate serial number: %w", err)
	}

	clientTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"GT Proxy Client"},
			CommonName:   clientName,
		},
		DNSNames:    []string{clientName},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(30 * 24 * time.Hour), // 30 days
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{ x509.ExtKeyUsageClientAuth },
	}

	// Add scopes as custom extension (OID 1.3.6.1.4.1.99999.1)
	if len(scopes) > 0 {
		scopeData := []byte(strings.Join(scopes, ","))
		clientTemplate.ExtraExtensions = []pkix.Extension{
			{
				Id:    asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 99999, 1},
				Value: scopeData,
			},
		}
	}

	// Sign client certificate with CA
	clientCertBytes, err := x509.CreateCertificate(rand.Reader, &clientTemplate, caCert, &clientKey.PublicKey, caKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to create client certificate: %w", err)
	}

	// Encode certificate to PEM
	certBuf := new(bytes.Buffer)
	if err := pem.Encode(certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: clientCertBytes}); err != nil {
		return "", "", fmt.Errorf("failed to encode certificate: %w", err)
	}

	// Encode key to PEM
	keyBytes, err := x509.MarshalECPrivateKey(clientKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal client key: %w", err)
	}

	keyBuf := new(bytes.Buffer)
	if err := pem.Encode(keyBuf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}); err != nil {
		return "", "", fmt.Errorf("failed to encode key: %w", err)
	}

	return certBuf.String(), keyBuf.String(), nil
}

// SaveClientCert saves a client certificate to disk
func (m *Manager) SaveClientCert(clientName, certPEM, keyPEM string) (certPath, keyPath string, err error) {
	clientDir := filepath.Join(m.certDir, "clients")
	if err := os.MkdirAll(clientDir, 0700); err != nil {
		return "", "", fmt.Errorf("failed to create client cert directory: %w", err)
	}

	certPath = filepath.Join(clientDir, fmt.Sprintf("%s.crt", clientName))
	keyPath = filepath.Join(clientDir, fmt.Sprintf("%s.key", clientName))

	if err := os.WriteFile(certPath, []byte(certPEM), 0644); err != nil {
		return "", "", fmt.Errorf("failed to write client certificate: %w", err)
	}

	if err := os.WriteFile(keyPath, []byte(keyPEM), 0600); err != nil {
		return "", "", fmt.Errorf("failed to write client key: %w", err)
	}

	return certPath, keyPath, nil
}

// LoadClientCert loads a client certificate from disk
func (m *Manager) LoadClientCert(clientName string) (certPath, keyPath string, err error) {
	clientDir := filepath.Join(m.certDir, "clients")
	certPath = filepath.Join(clientDir, fmt.Sprintf("%s.crt", clientName))
	keyPath = filepath.Join(clientDir, fmt.Sprintf("%s.key", clientName))

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("client certificate not found: %s", clientName)
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("client key not found: %s", clientName)
	}

	return certPath, keyPath, nil
}

// ListClientCerts returns a list of all client certificate names
func (m *Manager) ListClientCerts() ([]string, error) {
	clientDir := filepath.Join(m.certDir, "clients")
	entries, err := os.ReadDir(clientDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	clients := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".crt" {
			name := strings.TrimSuffix(entry.Name(), ".crt")
			clients = append(clients, name)
		}
	}

	return clients, nil
}


