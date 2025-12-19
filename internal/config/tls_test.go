package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestServerTLSConfigLoadsCert(t *testing.T) {
	cert, key, err := GenerateSelfSigned("127.0.0.1", time.Hour)
	if err != nil {
		t.Fatalf("generate self-signed: %v", err)
	}
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	if err := os.WriteFile(certPath, cert, 0o644); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	if err := os.WriteFile(keyPath, key, 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}

	cfg := TLSConfig{Enabled: true, CertFile: certPath, KeyFile: keyPath}
	tlsCfg, err := ServerTLSConfig(cfg)
	if err != nil {
		t.Fatalf("server tls config: %v", err)
	}
	if len(tlsCfg.Certificates) == 0 {
		t.Fatalf("expected certificates loaded")
	}
}

func TestClientTLSConfigWithFingerprint(t *testing.T) {
	cert, _, err := GenerateSelfSigned("127.0.0.1", time.Hour)
	if err != nil {
		t.Fatalf("generate self-signed: %v", err)
	}
	sum := sha256.Sum256(certDER(t, cert))
	fp := hex.EncodeToString(sum[:])

	cfg := TLSConfig{
		Enabled:            true,
		InsecureSkipVerify: true,
		CertFingerprint:    fp,
	}
	tlsCfg, err := ClientTLSConfig(cfg)
	if err != nil {
		t.Fatalf("client tls config: %v", err)
	}
	if tlsCfg.VerifyPeerCertificate == nil {
		t.Fatalf("expected verify peer certificate set")
	}
}

func certDER(t *testing.T, pemBytes []byte) []byte {
	t.Helper()
	block, _ := pemDecode(pemBytes)
	if block == nil {
		t.Fatalf("failed to decode pem")
	}
	return block.Bytes
}

func pemDecode(b []byte) (*pem.Block, []byte) {
	return pem.Decode(b)
}
