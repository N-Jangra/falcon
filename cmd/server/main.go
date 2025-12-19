package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/njangra/falcon-tunnel/internal/config"
	"github.com/njangra/falcon-tunnel/internal/logger"
	"github.com/njangra/falcon-tunnel/internal/tunnel"
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

	var ln net.Listener
	if cfg.TLS.Enabled {
		tlsCfg, err := config.ServerTLSConfig(cfg.TLS)
		if err != nil {
			log.Fatalf("tls config: %v", err)
		}
		ln, err = tls.Listen("tcp", cfg.Server.ListenAddr, tlsCfg)
		if err != nil {
			log.Fatalf("tls listen: %v", err)
		}
	} else {
		var err error
		ln, err = net.Listen("tcp", cfg.Server.ListenAddr)
		if err != nil {
			log.Fatalf("listen: %v", err)
		}
	}
	defer ln.Close()

	l.Infof("listening on %s", ln.Addr())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := tunnel.NewServer(*cfg, nil, l)

	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	if err := server.Serve(ctx, ln); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
