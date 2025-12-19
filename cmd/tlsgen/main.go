package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/njangra/falcon-tunnel/internal/config"
)

func main() {
	host := flag.String("host", "localhost", "Comma-separated hostnames or IPs for the certificate")
	days := flag.Int("days", 365, "Certificate validity in days")
	certPath := flag.String("cert", "cert.pem", "Output certificate path")
	keyPath := flag.String("key", "key.pem", "Output private key path")
	flag.Parse()

	cert, key, err := config.GenerateSelfSigned(*host, time.Duration(*days)*24*time.Hour)
	if err != nil {
		log.Fatalf("generate self-signed cert: %v", err)
	}

	if err := os.WriteFile(*certPath, cert, 0o644); err != nil {
		log.Fatalf("write cert: %v", err)
	}
	if err := os.WriteFile(*keyPath, key, 0o600); err != nil {
		log.Fatalf("write key: %v", err)
	}

	fmt.Printf("Wrote cert: %s\nWrote key: %s\nHosts: %s\nValid: %d days\n", *certPath, *keyPath, *host, *days)
}
