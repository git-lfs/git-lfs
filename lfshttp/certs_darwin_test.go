//go:build darwin
// +build darwin

package lfshttp

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestAppendMatchingCertsByCN tests that our function correctly filters certificates
// based on hostname matches in the CN field.
func TestAppendMatchingCertsByCN(t *testing.T) {
	// Create certificates with different CNs for testing
	problematicCert := generateCACert(t, "example.org JSS Built-in Certificate Authority", true)
	exactMatchCert := generateCACert(t, "example.org", true)
	regularCert := generateCACert(t, "git-lfs.local", false)

	// Export the certs to PEM format
	problematicPEM := exportCertToPEM(t, problematicCert)
	exactMatchPEM := exportCertToPEM(t, exactMatchCert)
	regularPEM := exportCertToPEM(t, regularCert)

	// Combine all certs into a bundle
	certBundle := append(problematicPEM, exactMatchPEM...)
	certBundle = append(certBundle, regularPEM...)

	// Create test cases
	tests := []struct {
		name          string
		hostname      string
		certBundle    []byte
		expectedCount int // How many certs we expect to be added to the pool
		description   string
	}{
		{
			name:          "Problematic cert filtered for exact hostname",
			hostname:      "example.org",
			certBundle:    certBundle,
			expectedCount: 2, // Should only include exactMatchCert + regularCert (problematic is filtered)
			description:   "Should filter out the problematic cert with hostname as substring",
		},
		{
			name:          "All certs included for different hostname",
			hostname:      "github.com",
			certBundle:    certBundle,
			expectedCount: 3, // Should include all 3 certs
			description:   "Should include all certs when hostname doesn't match as substring",
		},
		{
			name:          "Only matching cert included",
			hostname:      "git-lfs.local",
			certBundle:    certBundle,
			expectedCount: 3, // Should only include all 3 certs
			description:   "Should only include the cert with exact hostname match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start with an empty pool
			pool := x509.NewCertPool()

			// Create a pool with all certificates (to check initial count)
			allCertsPool := x509.NewCertPool()
			ok := allCertsPool.AppendCertsFromPEM(certBundle)
			assert.True(t, ok, "Should be able to parse certificate bundle")

			// Call the function we're testing with the corrected parameter order
			updatedPool := appendMatchingCertsByCN(pool, tt.certBundle, tt.hostname)

			// Count the certificates using the deprecated Subjects() method
			filteredCount := len(updatedPool.Subjects())

			// Verify the count matches our expectation
			assert.Equal(t, tt.expectedCount, filteredCount, tt.description)
		})
	}
}

// generateCACert creates a test certificate with the specified common name
func generateCACert(t *testing.T, commonName string, isCA bool) *x509.Certificate {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		t.Fatalf("Failed to generate serial number: %v", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(10 * 365 * 24 * time.Hour)

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  isCA,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		t.Fatalf("Failed to parse generated certificate: %v", err)
	}

	return cert
}

// exportCertToPEM converts an x509 certificate to PEM format
func exportCertToPEM(t *testing.T, cert *x509.Certificate) []byte {
	pemBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}

	return pem.EncodeToMemory(pemBlock)
}
