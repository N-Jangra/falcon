package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/njangra/falcon-tunnel/internal/config"
)

func TestSetupWithFileOutput(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "app.log")

	l, cleanup, err := Setup(config.LogConfig{
		Level:    "info",
		FilePath: logPath,
		Format:   "text",
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if cleanup == nil {
		t.Fatalf("expected cleanup when file output is configured")
	}
	l.Info("hello world")
	if err := cleanup(); err != nil {
		t.Fatalf("cleanup: %v", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(data), "hello world") {
		t.Fatalf("expected log message written to file")
	}
}

func TestSetupValidatesLevelAndFormat(t *testing.T) {
	_, _, err := Setup(config.LogConfig{Level: "bad-level"})
	if err == nil {
		t.Fatalf("expected error on invalid level")
	}

	_, _, err = Setup(config.LogConfig{Level: "info", Format: "xml"})
	if err == nil {
		t.Fatalf("expected error on invalid format")
	}
}
