package logger

import (
	"context"
	"time"

	"trade/internal/ports"

	"github.com/sirupsen/logrus"
)

type LogrusAdapter struct {
	log *logrus.Logger
}

func NewLogrusAdapter(levelStr string) (*LogrusAdapter, error) {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339})
	lvl, err := logrus.ParseLevel(levelStr)
	if err != nil {
		return nil, err
	}
	l.SetLevel(lvl)
	return &LogrusAdapter{log: l}, nil
}

func (l *LogrusAdapter) Info(ctx context.Context, msg string, fields ports.Fields) {
	entry := l.log.WithFields(logrus.Fields(fields))
	entry.Info(msg)
}

func (l *LogrusAdapter) Error(ctx context.Context, msg string, fields ports.Fields) {
	entry := l.log.WithFields(logrus.Fields(fields))
	entry.Error(msg)
}

func (l *LogrusAdapter) Debug(ctx context.Context, msg string, fields ports.Fields) {
	entry := l.log.WithFields(logrus.Fields(fields))
	entry.Debug(msg)
}
