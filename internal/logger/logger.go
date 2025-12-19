package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/njangra/falcon-tunnel/internal/config"
	"github.com/sirupsen/logrus"
)

// Setup builds a logrus Logger according to configuration.
// It returns an optional cleanup function to close any file handles.
func Setup(cfg config.LogConfig) (*logrus.Logger, func() error, error) {
	l := logrus.New()

	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return nil, nil, fmt.Errorf("parse log level: %w", err)
	}
	l.SetLevel(level)

	switch cfg.Format {
	case "json":
		l.SetFormatter(&logrus.JSONFormatter{})
	case "text", "":
		l.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	default:
		return nil, nil, fmt.Errorf("invalid log format: %s", cfg.Format)
	}

	var writers []io.Writer
	writers = append(writers, os.Stdout)

	var closeFn func() error
	if cfg.FilePath != "" {
		f, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, nil, fmt.Errorf("open log file: %w", err)
		}
		writers = append(writers, f)
		closeFn = f.Close
	}

	l.SetOutput(io.MultiWriter(writers...))
	return l, closeFn, nil
}
