// Copyright (c) 2016 Company 0, LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package server

// newTLSCertPair returns a new PEM-encoded x.509 certificate pair based on a
import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// generateSelfSignedCert generates a new ECDSA private key and self-signed certificate.
func generateSelfSignedCert(organization string, validUntil time.Time, extraHosts []string) ([]byte, []byte, error) {
	// Generate a new private key
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	// Create a certificate template
	notBefore := time.Now()
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{organization},
		},
		NotBefore:             notBefore,
		NotAfter:              validUntil,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	for _, host := range extraHosts {
		if ip := net.ParseIP(host); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, host)
		}
	}

	// Generate the self-signed certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %v", err)
	}

	// Encode the certificate to PEM format
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// Encode the private key to PEM format
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal private key: %v", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	return certPEM, keyPEM, nil
}

// GenerateNewTLSCertPair generates a new TLS certificate and key pair and saves them to the specified paths.
func (s *Server) GenerateNewTLSCertPair(organization string, validUntil time.Time, extraHosts []string, certPath, keyPath string) error {
	// Generate the certificate and key
	cert, key, err := generateSelfSignedCert(organization, validUntil, extraHosts)
	if err != nil {
		return fmt.Errorf("failed to generate new TLS certificate pair: %v", err)
	}

	// Ensure the directory for the certificate and key exists
	if err := os.MkdirAll(filepath.Dir(certPath), 0700); err != nil {
		return fmt.Errorf("failed to create directory for cert files: %v", err)
	}

	// Write the certificate to file
	if err := os.WriteFile(certPath, cert, 0600); err != nil {
		return fmt.Errorf("failed to write certificate to file: %v", err)
	}

	// Write the key to file
	if err := os.WriteFile(keyPath, key, 0600); err != nil {
		return fmt.Errorf("failed to write key to file: %v", err)
	}

	fmt.Printf("New TLS certificate and key pair generated and saved to %s and %s\n", certPath, keyPath)
	return nil
}
