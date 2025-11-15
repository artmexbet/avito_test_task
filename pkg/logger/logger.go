package logger

import (
	"log/slog"
	"os"
)

type EnvType string

const (
	EnvDevelopment EnvType = "development"
	EnvProduction  EnvType = "production"
)

func NewLogger(envType EnvType) *slog.Logger {
	switch envType {
	case EnvDevelopment:
		return slog.New(slog.NewTextHandler(os.Stdout,
			&slog.HandlerOptions{ //nolint:exhaustruct
				AddSource: true,
				Level:     slog.LevelDebug,
			}))
	case EnvProduction:
		return slog.New(slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{ //nolint:exhaustruct
				Level: slog.LevelInfo,
			}))
	}
	return slog.New(slog.NewTextHandler(os.Stdout,
		&slog.HandlerOptions{ //nolint:exhaustruct
			AddSource: true,
			Level:     slog.LevelInfo,
		}))
}
