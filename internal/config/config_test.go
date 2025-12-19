package config

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidateAndDefaults(t *testing.T) {
	yaml := `
server:
  listen_addr: ":9090"
  ftp_server_addr: "ftp.internal:21"
client:
  tunnel_addr: "server.internal:8080"
  local_ftp_port: 2021
auth:
  enabled: true
  password_hash: "$2a$10$abcdefghijklmnopqrstuv"
tls:
  enabled: true
  cert_file: "cert.pem"
  key_file: "key.pem"
log:
  level: "debug"
`
	cfg, err := Load([]byte(yaml))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	ApplyDefaults(cfg)
	if err := Validate(cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if cfg.Server.Timeout == 0 || cfg.Client.Timeout == 0 {
		t.Fatalf("expected default timeouts applied")
	}
	if cfg.Log.Format != "text" {
		t.Fatalf("expected default log format")
	}
}

func TestApplyOverridesHashesPassword(t *testing.T) {
	cfg := Default()
	cfg.Server.FTPServerAddr = "ftp.internal:21"
	cfg.Client.TunnelAddr = "server:8080"

	overridePassword := "supersecret"
	overrides := Overrides{
		Password:       &overridePassword,
		MaxConnections: intPtr(50),
		AuthEnabled:    boolPtr(true),
		FTPServerAddr:  strPtr("ftp.override:21"),
	}

	if err := ApplyOverrides(&cfg, overrides); err != nil {
		t.Fatalf("apply overrides: %v", err)
	}
	if cfg.Server.MaxConnections != 50 {
		t.Fatalf("expected max connections override")
	}
	if cfg.Server.FTPServerAddr != "ftp.override:21" {
		t.Fatalf("expected ftp address override")
	}
	if cfg.Auth.PasswordHash == "" {
		t.Fatalf("expected password hash to be set")
	}
	if err := Validate(&cfg); err != nil {
		t.Fatalf("validate after overrides: %v", err)
	}
}

func TestValidateFailures(t *testing.T) {
	cfg := Default()
	if err := Validate(&cfg); err == nil {
		t.Fatalf("expected validation to fail due to missing fields")
	}

	cfg.Server.FTPServerAddr = "ftp:21"
	cfg.Client.TunnelAddr = "server:8080"
	cfg.Client.LocalFTPPort = 0
	if err := Validate(&cfg); err != ErrMissingLocalFTPPort {
		t.Fatalf("expected missing local ftp port error, got %v", err)
	}

	cfg.Client.LocalFTPPort = 2021
	cfg.Auth.Enabled = true
	if err := Validate(&cfg); err != ErrMissingPasswordHash {
		t.Fatalf("expected missing password hash error, got %v", err)
	}
}

func TestBuildWithFileAndOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
server:
  listen_addr: ":8081"
  ftp_server_addr: "ftp.sample:21"
  max_connections: 20
client:
  tunnel_addr: "server.sample:8080"
  local_ftp_port: 2021
auth:
  enabled: false
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	overrideLog := "warn"
	cfg, err := Build(path, Overrides{LogLevel: &overrideLog})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if cfg.Log.Level != "warn" {
		t.Fatalf("expected log level override applied")
	}
	if cfg.Server.Timeout == 0 || cfg.Client.Timeout == 0 {
		t.Fatalf("expected defaults applied")
	}
}

func TestCLIFlagOverrides(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	flags := RegisterFlags(fs)
	args := []string{
		"--listen", ":9000",
		"--ftp", "ftp.flag:21",
		"--server", "srv:8080",
		"--local-port", "2022",
		"--auth=true",
		"--password", "mypassword",
		"--log-level", "error",
	}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("parse flags: %v", err)
	}
	ov := OverridesFromFlags(flags)

	cfg := Default()
	if err := ApplyOverrides(&cfg, ov); err != nil {
		t.Fatalf("apply overrides: %v", err)
	}
	cfg.Server.FTPServerAddr = cfg.Server.FTPServerAddr // keep set from overrides
	ApplyDefaults(&cfg)
	if err := Validate(&cfg); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if cfg.Log.Level != "error" {
		t.Fatalf("expected log level to be error")
	}
	if cfg.Auth.PasswordHash == "" {
		t.Fatalf("expected password to be hashed from flag")
	}
}

func intPtr(v int) *int       { return &v }
func boolPtr(v bool) *bool    { return &v }
func strPtr(v string) *string { return &v }
