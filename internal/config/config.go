package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/njangra/falcon-tunnel/internal/auth"
	"gopkg.in/yaml.v3"
)

// Config is the root application configuration structure.
type Config struct {
	Server ServerConfig `yaml:"server"`
	Client ClientConfig `yaml:"client"`
	Auth   AuthConfig   `yaml:"auth"`
	TLS    TLSConfig    `yaml:"tls"`
	Log    LogConfig    `yaml:"log"`
}

type ServerConfig struct {
	ListenAddr     string        `yaml:"listen_addr"`
	FTPServerAddr  string        `yaml:"ftp_server_addr"`
	MaxConnections int           `yaml:"max_connections"`
	Timeout        time.Duration `yaml:"timeout"`
}

type ClientConfig struct {
	TunnelAddr   string        `yaml:"tunnel_addr"`
	LocalFTPPort int           `yaml:"local_ftp_port"`
	Timeout      time.Duration `yaml:"timeout"`
	Password     string        `yaml:"password"` // plaintext password to send when auth is enabled
}

type AuthConfig struct {
	Enabled      bool   `yaml:"enabled"`
	PasswordHash string `yaml:"password_hash"`
}

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type LogConfig struct {
	Level    string `yaml:"level"`
	FilePath string `yaml:"file_path"`
	Format   string `yaml:"format"` // text or json
}

// Load parses YAML bytes into Config.
func Load(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoadFile reads YAML configuration from a file path.
func LoadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Load(data)
}

// Default returns a Config populated with reasonable defaults.
func Default() Config {
	return Config{
		Server: ServerConfig{
			ListenAddr:     ":8080",
			MaxConnections: 100,
			Timeout:        30 * time.Second,
		},
		Client: ClientConfig{
			LocalFTPPort: 2121,
			Timeout:      30 * time.Second,
		},
		Auth: AuthConfig{
			Enabled:      false,
			PasswordHash: "",
		},
		TLS: TLSConfig{
			Enabled: false,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

// ApplyDefaults sets defaults on a Config in-place where values are zero.
func ApplyDefaults(cfg *Config) {
	defaults := Default()

	if cfg.Server.ListenAddr == "" {
		cfg.Server.ListenAddr = defaults.Server.ListenAddr
	}
	if cfg.Server.MaxConnections == 0 {
		cfg.Server.MaxConnections = defaults.Server.MaxConnections
	}
	if cfg.Server.Timeout == 0 {
		cfg.Server.Timeout = defaults.Server.Timeout
	}

	if cfg.Client.LocalFTPPort == 0 {
		cfg.Client.LocalFTPPort = defaults.Client.LocalFTPPort
	}
	if cfg.Client.Timeout == 0 {
		cfg.Client.Timeout = defaults.Client.Timeout
	}

	if cfg.Log.Level == "" {
		cfg.Log.Level = defaults.Log.Level
	}
	if cfg.Log.Format == "" {
		cfg.Log.Format = defaults.Log.Format
	}
}

// Validation errors for required fields.
var (
	ErrMissingServerListenAddr = errors.New("server.listen_addr is required")
	ErrMissingFTPServerAddr    = errors.New("server.ftp_server_addr is required")
	ErrMissingTunnelAddr       = errors.New("client.tunnel_addr is required")
	ErrMissingLocalFTPPort     = errors.New("client.local_ftp_port must be > 0")
	ErrMissingClientPassword   = errors.New("client.password is required when auth is enabled")
	ErrMissingPasswordHash     = errors.New("auth.password_hash is required when auth is enabled")
	ErrMissingTLSCert          = errors.New("tls.cert_file is required when TLS is enabled")
	ErrMissingTLSKey           = errors.New("tls.key_file is required when TLS is enabled")
	ErrInvalidMaxConnections   = errors.New("server.max_connections must be > 0")
	ErrInvalidTimeout          = errors.New("timeout must be > 0")
)

// Validate checks required and minimal values.
func Validate(cfg *Config) error {
	if cfg.Server.ListenAddr == "" {
		return ErrMissingServerListenAddr
	}
	if cfg.Server.FTPServerAddr == "" {
		return ErrMissingFTPServerAddr
	}
	if cfg.Server.MaxConnections <= 0 {
		return ErrInvalidMaxConnections
	}
	if cfg.Server.Timeout <= 0 {
		return ErrInvalidTimeout
	}
	if cfg.Client.TunnelAddr == "" {
		return ErrMissingTunnelAddr
	}
	if cfg.Client.LocalFTPPort <= 0 {
		return ErrMissingLocalFTPPort
	}
	if cfg.Client.Timeout <= 0 {
		return ErrInvalidTimeout
	}
	if cfg.Auth.Enabled && cfg.Auth.PasswordHash == "" {
		return ErrMissingPasswordHash
	}
	if cfg.Auth.Enabled && cfg.Client.Password == "" {
		return ErrMissingClientPassword
	}
	if cfg.TLS.Enabled {
		if cfg.TLS.CertFile == "" {
			return ErrMissingTLSCert
		}
		if cfg.TLS.KeyFile == "" {
			return ErrMissingTLSKey
		}
	}
	return nil
}

// Overrides hold optional CLI-provided overrides.
type Overrides struct {
	ConfigPath     *string
	ListenAddr     *string
	FTPServerAddr  *string
	MaxConnections *int
	ServerTimeout  *time.Duration
	TunnelAddr     *string
	LocalFTPPort   *int
	ClientTimeout  *time.Duration
	ClientPassword *string
	AuthEnabled    *bool
	Password       *string
	PasswordHash   *string
	TLSEnabled     *bool
	TLSCertFile    *string
	TLSKeyFile     *string
	LogLevel       *string
	LogFilePath    *string
	LogFormat      *string
}

// ApplyOverrides mutates cfg using non-nil override values.
func ApplyOverrides(cfg *Config, o Overrides) error {
	if o.ListenAddr != nil {
		cfg.Server.ListenAddr = *o.ListenAddr
	}
	if o.FTPServerAddr != nil {
		cfg.Server.FTPServerAddr = *o.FTPServerAddr
	}
	if o.MaxConnections != nil {
		cfg.Server.MaxConnections = *o.MaxConnections
	}
	if o.ServerTimeout != nil {
		cfg.Server.Timeout = *o.ServerTimeout
	}
	if o.TunnelAddr != nil {
		cfg.Client.TunnelAddr = *o.TunnelAddr
	}
	if o.LocalFTPPort != nil {
		cfg.Client.LocalFTPPort = *o.LocalFTPPort
	}
	if o.ClientTimeout != nil {
		cfg.Client.Timeout = *o.ClientTimeout
	}
	if o.ClientPassword != nil {
		cfg.Client.Password = *o.ClientPassword
	}
	if o.AuthEnabled != nil {
		cfg.Auth.Enabled = *o.AuthEnabled
	}
	if o.PasswordHash != nil {
		cfg.Auth.PasswordHash = *o.PasswordHash
	}
	if o.Password != nil {
		hash, err := auth.HashPassword(*o.Password)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}
		cfg.Auth.PasswordHash = hash
	}
	if o.TLSEnabled != nil {
		cfg.TLS.Enabled = *o.TLSEnabled
	}
	if o.TLSCertFile != nil {
		cfg.TLS.CertFile = *o.TLSCertFile
	}
	if o.TLSKeyFile != nil {
		cfg.TLS.KeyFile = *o.TLSKeyFile
	}
	if o.LogLevel != nil {
		cfg.Log.Level = *o.LogLevel
	}
	if o.LogFilePath != nil {
		cfg.Log.FilePath = *o.LogFilePath
	}
	if o.LogFormat != nil {
		cfg.Log.Format = *o.LogFormat
	}
	return nil
}

// Flag wrappers to track whether a flag was set.
type stringFlag struct {
	value string
	set   bool
}

func (f *stringFlag) String() string { return f.value }
func (f *stringFlag) Set(v string) error {
	f.value = v
	f.set = true
	return nil
}

type intFlag struct {
	value int
	set   bool
}

func (f *intFlag) String() string { return fmt.Sprintf("%d", f.value) }
func (f *intFlag) Set(v string) error {
	var parsed int
	_, err := fmt.Sscanf(v, "%d", &parsed)
	if err != nil {
		return err
	}
	f.value = parsed
	f.set = true
	return nil
}

type durationFlag struct {
	value time.Duration
	set   bool
}

func (f *durationFlag) String() string { return f.value.String() }
func (f *durationFlag) Set(v string) error {
	d, err := time.ParseDuration(v)
	if err != nil {
		return err
	}
	f.value = d
	f.set = true
	return nil
}

type boolFlag struct {
	value bool
	set   bool
}

func (f *boolFlag) String() string { return fmt.Sprintf("%t", f.value) }
func (f *boolFlag) Set(v string) error {
	if v == "" {
		f.value = true
		f.set = true
		return nil
	}
	switch v {
	case "true", "1":
		f.value = true
	case "false", "0":
		f.value = false
	default:
		return fmt.Errorf("invalid bool %q", v)
	}
	f.set = true
	return nil
}

// CLIFlags holds registered flag pointers for reuse in both binaries.
type CLIFlags struct {
	ConfigPath stringFlag

	ListenAddr     stringFlag
	FTPServerAddr  stringFlag
	MaxConnections intFlag
	ServerTimeout  durationFlag

	TunnelAddr     stringFlag
	LocalFTPPort   intFlag
	ClientTimeout  durationFlag
	ClientPassword stringFlag

	AuthEnabled  boolFlag
	Password     stringFlag
	PasswordHash stringFlag

	TLSEnabled  boolFlag
	TLSCertFile stringFlag
	TLSKeyFile  stringFlag

	LogLevel    stringFlag
	LogFilePath stringFlag
	LogFormat   stringFlag
}

// RegisterFlags binds CLI flags on the provided FlagSet.
func RegisterFlags(fs *flag.FlagSet) *CLIFlags {
	flags := &CLIFlags{}

	fs.Var(&flags.ConfigPath, "config", "Path to YAML configuration file")

	fs.Var(&flags.ListenAddr, "listen", "Server listen address (e.g. :8080)")
	fs.Var(&flags.FTPServerAddr, "ftp", "Target FTP server address (host:port)")
	fs.Var(&flags.MaxConnections, "max-conns", "Maximum concurrent connections")
	fs.Var(&flags.ServerTimeout, "server-timeout", "Server timeout (e.g. 30s)")

	fs.Var(&flags.TunnelAddr, "server", "Tunnel server address (host:port)")
	fs.Var(&flags.LocalFTPPort, "local-port", "Local FTP port to listen on")
	fs.Var(&flags.ClientTimeout, "client-timeout", "Client timeout (e.g. 30s)")
	fs.Var(&flags.ClientPassword, "client-password", "Plaintext password for client authentication")

	fs.Var(&flags.AuthEnabled, "auth", "Enable authentication (true/false)")
	fs.Var(&flags.Password, "password", "Plaintext password (hashed internally)")
	fs.Var(&flags.PasswordHash, "password-hash", "Existing bcrypt password hash")

	fs.Var(&flags.TLSEnabled, "tls", "Enable TLS (true/false)")
	fs.Var(&flags.TLSCertFile, "tls-cert", "TLS certificate file")
	fs.Var(&flags.TLSKeyFile, "tls-key", "TLS private key file")

	fs.Var(&flags.LogLevel, "log-level", "Log level (debug, info, warn, error)")
	fs.Var(&flags.LogFilePath, "log-file", "Log file path (optional)")
	fs.Var(&flags.LogFormat, "log-format", "Log format: text or json")

	return flags
}

// OverridesFromFlags converts parsed CLIFlags into Overrides.
func OverridesFromFlags(f *CLIFlags) Overrides {
	ov := Overrides{}

	if f.ConfigPath.set {
		ov.ConfigPath = &f.ConfigPath.value
	}
	if f.ListenAddr.set {
		ov.ListenAddr = &f.ListenAddr.value
	}
	if f.FTPServerAddr.set {
		ov.FTPServerAddr = &f.FTPServerAddr.value
	}
	if f.MaxConnections.set {
		ov.MaxConnections = &f.MaxConnections.value
	}
	if f.ServerTimeout.set {
		ov.ServerTimeout = &f.ServerTimeout.value
	}
	if f.TunnelAddr.set {
		ov.TunnelAddr = &f.TunnelAddr.value
	}
	if f.LocalFTPPort.set {
		ov.LocalFTPPort = &f.LocalFTPPort.value
	}
	if f.ClientTimeout.set {
		ov.ClientTimeout = &f.ClientTimeout.value
	}
	if f.ClientPassword.set {
		ov.ClientPassword = &f.ClientPassword.value
	}
	if f.AuthEnabled.set {
		ov.AuthEnabled = &f.AuthEnabled.value
	}
	if f.Password.set {
		ov.Password = &f.Password.value
	}
	if f.PasswordHash.set {
		ov.PasswordHash = &f.PasswordHash.value
	}
	if f.TLSEnabled.set {
		ov.TLSEnabled = &f.TLSEnabled.value
	}
	if f.TLSCertFile.set {
		ov.TLSCertFile = &f.TLSCertFile.value
	}
	if f.TLSKeyFile.set {
		ov.TLSKeyFile = &f.TLSKeyFile.value
	}
	if f.LogLevel.set {
		ov.LogLevel = &f.LogLevel.value
	}
	if f.LogFilePath.set {
		ov.LogFilePath = &f.LogFilePath.value
	}
	if f.LogFormat.set {
		ov.LogFormat = &f.LogFormat.value
	}
	return ov
}

// Build constructs a Config using defaults, optional file, and overrides.
func Build(filePath string, overrides Overrides) (*Config, error) {
	var cfg Config
	if filePath != "" {
		loaded, err := LoadFile(filePath)
		if err != nil {
			return nil, err
		}
		cfg = *loaded
	} else {
		cfg = Default()
	}

	if err := ApplyOverrides(&cfg, overrides); err != nil {
		return nil, err
	}
	ApplyDefaults(&cfg)
	if err := Validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
