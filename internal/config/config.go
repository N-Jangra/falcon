package config

import (
	"time"

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
	Format   string `yaml:"format"`
}

// Load parses YAML bytes into Config.
func Load(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
