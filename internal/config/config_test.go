package config

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadValidateAndDefaults(t *testing.T) {
	yaml := `
server:
  listen_addr: ":9090"
  ftp_server_addr: "ftp.internal:21"
client:
  tunnel_addr: "server.internal:8080"
  local_ftp_port: 2021
  password: "secret"
  idle_timeout: 5s
  keepalive: 2s
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
	if cfg.Client.IdleTimeout != 5*time.Second {
		t.Fatalf("expected client idle timeout parsed")
	}
}

func TestApplyOverridesHashesPassword(t *testing.T) {
	cfg := Default()
	cfg.Server.FTPServerAddr = "ftp.internal:21"
	cfg.Client.TunnelAddr = "server:8080"
	cfg.Client.Password = "pw"

	overridePassword := "supersecret"
	overrides := Overrides{
		Password:             &overridePassword,
		MaxConnections:       intPtr(50),
		AuthEnabled:          boolPtr(true),
		FTPServerAddr:        strPtr("ftp.override:21"),
		ClientPassword:       strPtr("clientpass"),
		ClientIdle:           durationPtr(10 * time.Second),
		ClientKeepAlive:      durationPtr(5 * time.Second),
		ClientRetries:        intPtr(5),
		ClientBackoffInitial: durationPtr(200 * time.Millisecond),
		ClientBackoffMax:     durationPtr(3 * time.Second),
		ServerIdle:           durationPtr(45 * time.Second),
		PoolSize:             intPtr(10),
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
	if cfg.Client.Password != "clientpass" {
		t.Fatalf("expected client password override applied")
	}
	if cfg.Client.IdleTimeout != 10*time.Second || cfg.Client.KeepAlive != 5*time.Second {
		t.Fatalf("expected client idle/keepalive overrides applied")
	}
	if cfg.Client.MaxRetries != 5 || cfg.Client.BackoffInitial != 200*time.Millisecond || cfg.Client.BackoffMax != 3*time.Second {
		t.Fatalf("expected client backoff overrides applied")
	}
	if cfg.Server.IdleTimeout != 45*time.Second || cfg.Server.PoolSize != 10 {
		t.Fatalf("expected server idle timeout and pool size overrides applied")
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
	cfg.Auth.PasswordHash = "hash"
	if err := Validate(&cfg); err != ErrMissingClientPassword {
		t.Fatalf("expected missing client password error, got %v", err)
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
  password: "pw"
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
		"--client-password", "clientpw",
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

func intPtr(v int) *int                          { return &v }
func boolPtr(v bool) *bool                       { return &v }
func strPtr(v string) *string                    { return &v }
func durationPtr(d time.Duration) *time.Duration { return &d }
