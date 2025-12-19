package config

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

// ServerTLSConfig builds a tls.Config for servers using the provided TLSConfig.
func ServerTLSConfig(cfg TLSConfig) (*tls.Config, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if cfg.CertFile == "" || cfg.KeyFile == "" {
		return nil, fmt.Errorf("tls cert_file and key_file are required")
	}
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load key pair: %w", err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// ClientTLSConfig builds a tls.Config for clients using the provided TLSConfig.
// CertFingerprint, when set, enforces SHA-256 fingerprint matching (hex string).
func ClientTLSConfig(cfg TLSConfig) (*tls.Config, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: cfg.ServerName,
	}

	if cfg.CAFile != "" {
		pool, err := loadCertPool(cfg.CAFile)
		if err != nil {
			return nil, err
		}
		tlsCfg.RootCAs = pool
	}

	if cfg.InsecureSkipVerify || cfg.CertFingerprint != "" {
		tlsCfg.InsecureSkipVerify = true
	}

	if cfg.CertFingerprint != "" {
		expect, err := parseFingerprint(cfg.CertFingerprint)
		if err != nil {
			return nil, err
		}
		tlsCfg.VerifyPeerCertificate = func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			if len(rawCerts) == 0 {
				return errors.New("no peer certificate presented")
			}
			sum := sha256.Sum256(rawCerts[0])
			if !bytes.Equal(sum[:], expect) {
				return fmt.Errorf("certificate fingerprint mismatch")
			}
			return nil
		}
	}

	return tlsCfg, nil
}

func loadCertPool(path string) (*x509.CertPool, error) {
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read ca file: %w", err)
	}
	if ok := pool.AppendCertsFromPEM(data); !ok {
		return nil, fmt.Errorf("failed to append CA certs from %s", path)
	}
	return pool, nil
}

func parseFingerprint(fp string) ([]byte, error) {
	fp = strings.ReplaceAll(fp, ":", "")
	fp = strings.ToLower(fp)
	b, err := hex.DecodeString(fp)
	if err != nil {
		return nil, fmt.Errorf("decode fingerprint: %w", err)
	}
	if len(b) != sha256.Size {
		return nil, fmt.Errorf("fingerprint must be %d bytes", sha256.Size)
	}
	return b, nil
}

// GenerateSelfSigned creates a self-signed certificate for the given host (IP or DNS).
// It returns PEM-encoded cert and key bytes.
func GenerateSelfSigned(host string, validFor time.Duration) ([]byte, []byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("generate key: %w", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: host,
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		h = strings.TrimSpace(h)
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else if h != "" {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("create certificate: %w", err)
	}

	certBuf := &bytes.Buffer{}
	keyBuf := &bytes.Buffer{}
	if err := pem.Encode(certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return nil, nil, fmt.Errorf("encode cert: %w", err)
	}
	if err := pem.Encode(keyBuf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}); err != nil {
		return nil, nil, fmt.Errorf("encode key: %w", err)
	}

	return certBuf.Bytes(), keyBuf.Bytes(), nil
}
