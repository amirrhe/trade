package ports

import "context"

type Fields map[string]interface{}

type LoggerPort interface {
	Info(ctx context.Context, msg string, fields Fields)
	Error(ctx context.Context, msg string, fields Fields)
	Debug(ctx context.Context, msg string, fields Fields)
}
