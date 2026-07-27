package logger

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/n0m-d/DVAPI/internal/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

var sensitiveKeys = []string{"password", "key", "secret", "otp"}

func New(cfg *config.Config) (*slog.Logger, func() error, error) {

	// Replace any potentially sensitive values with the string [REDACTED]
	replaceAttr := func(groups []string, a slog.Attr) slog.Attr {
		if slices.Contains(sensitiveKeys, a.Key) {
			return slog.String(a.Key, "[REDACTED]")
		}
		if a.Key == "query" && strings.Contains(a.Value.String(), "token=") { // redact token from query string for stream requests
			return slog.String(a.Key, "token=[REDACTED]")
		}

		return a
	}

	textHandler := tint.NewTextHandler(os.Stderr, &tint.Options{
		Level:       slog.LevelDebug,
		TimeFormat:  time.Kitchen,
		NoColor:     false,
		ReplaceAttr: replaceAttr,
	})

	if cfg.LogFile != "" {

		logger := &lumberjack.Logger{
			Filename:   cfg.LogFile,
			MaxSize:    1,
			MaxAge:     28,
			MaxBackups: 10,
			LocalTime:  false,
			Compress:   true,
		}

		logFile := cfg.LogFile
		debugHandler := textHandler

		// Create log directory if it doesn't exist
		dir := filepath.Dir(logFile)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Open log file, create if it doesn't exist
		file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open log file: %w", err)
		}

		// Buffer log file writes to improve performance
		bufferedFile := bufio.NewWriterSize(file, 8192)

		// Create info handler for JSON logging
		infoHandler := slog.NewJSONHandler(logger, &slog.HandlerOptions{
			Level:       slog.LevelInfo,
			ReplaceAttr: replaceAttr,
		})

		close := func() error {
			if err := bufferedFile.Flush(); err != nil {
				return fmt.Errorf("failed to flush log file: %w", err)
			}
			if err := logger.Close(); err != nil {
				return fmt.Errorf("failed to close log file: %w", err)
			}
			return nil
		}
		if cfg.Env == "development" {
			return slog.New(slog.NewMultiHandler(debugHandler, infoHandler)), close, nil
		} else {
			// In production, only log to file
			return slog.New(infoHandler), close, nil
		}
	}
	close := func() error {
		return nil
	}

	return slog.New(textHandler), close, nil
}
