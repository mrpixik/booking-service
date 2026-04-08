package sl

import (
	"log/slog"
	"os"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/adapters/logger"
	"github.com/avito-internships/test-backend-1-mrpixik/pkg/logger/sl/handlers/slogpretty"
)

type SlogLogger struct {
	l *slog.Logger
}

func NewSlogLogger(env string) *SlogLogger {
	switch env {
	case "local":
		return &SlogLogger{l: slogpretty.SetupPrettySlog()}
	case "dev":
		return &SlogLogger{slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))}
	case "prod":
		return &SlogLogger{slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))}
	default:
		return &SlogLogger{slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))}
	}
}

func (sl *SlogLogger) Info(msg string, args ...any) {
	sl.l.Info(msg, args...)
}

func (sl *SlogLogger) Error(msg string, args ...any) {
	sl.l.Error(msg, args...)
}

func (sl *SlogLogger) Warn(msg string, args ...any) {
	sl.l.Warn(msg, args...)
}

func (sl *SlogLogger) Debug(msg string, args ...any) {
	sl.l.Debug(msg, args...)
}

func (sl *SlogLogger) With(args ...any) logger.LoggerAdapter {
	return &SlogLogger{l: sl.l.With(args...)}
}
