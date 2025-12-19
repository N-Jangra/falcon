package logger

import "github.com/sirupsen/logrus"

// New creates a baseline structured logger.
func New() *logrus.Logger {
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	l.SetLevel(logrus.InfoLevel)
	return l
}
