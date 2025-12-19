package main

import (
	"flag"
	"log"
	"os"

	"github.com/njangra/falcon-tunnel/internal/config"
	"github.com/njangra/falcon-tunnel/internal/logger"
)

func main() {
	fs := flag.NewFlagSet("falcon-tunnel-server", flag.ExitOnError)
	flags := config.RegisterFlags(fs)
	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatalf("parse flags: %v", err)
	}

	ov := config.OverridesFromFlags(flags)
	configPath := ""
	if ov.ConfigPath != nil {
		configPath = *ov.ConfigPath
	}

	cfg, err := config.Build(configPath, ov)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	l, cleanup, err := logger.Setup(cfg.Log)
	if err != nil {
		log.Fatalf("logger setup: %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	l.WithFields(map[string]any{
		"listen": cfg.Server.ListenAddr,
		"ftp":    cfg.Server.FTPServerAddr,
		"tls":    cfg.TLS.Enabled,
	}).Info("server configuration loaded")

	// TODO: implement server start in later sprints.
}
