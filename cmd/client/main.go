package main

import (
	"context"
	"flag"
	"fmt"
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
	fs := flag.NewFlagSet("falcon-tunnel-client", flag.ExitOnError)
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
		"server": cfg.Client.TunnelAddr,
		"port":   cfg.Client.LocalFTPPort,
		"tls":    cfg.TLS.Enabled,
	}).Info("client configuration loaded")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := tunnel.NewClient(*cfg, l)

	// Close local listener when context cancels by connecting to it to unblock accept if needed.
	go func() {
		<-ctx.Done()
		_, _ = net.Dial("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", cfg.Client.LocalFTPPort)))
	}()

	if err := client.Start(ctx); err != nil {
		log.Fatalf("client error: %v", err)
	}
}
