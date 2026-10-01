// Package logger configures the application's structured logger.
//
// Every component receives a *zerolog.Logger explicitly (no global
// logger) so call sites stay testable and log output stays predictable.
package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
)

// New builds a zerolog.Logger configured for the given app environment.
// Production and staging emit compact JSON (suitable for log aggregators);
// development emits a human-readable console format.
func New(cfg config.AppConfig) zerolog.Logger {
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = time.RFC3339Nano

	var writer = os.Stdout

	base := zerolog.New(writer).With().Timestamp()

	if !cfg.IsProduction() {
		console := zerolog.ConsoleWriter{Out: writer, TimeFormat: time.Kitchen}
		return zerolog.New(console).With().Timestamp().Str("env", cfg.Env).Logger()
	}

	return base.Str("env", cfg.Env).Logger()
}
