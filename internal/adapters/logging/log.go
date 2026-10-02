package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"uuid"
)

const (
	dirPermissions  = 0o700
	filePermissions = 0o600
	appendFlags     = os.O_WRONLY | os.O_CREATE | os.O_APPEND
)

type Log struct {
	file   *os.File
	logger *slog.Logger
}

func Open(path string) (*Log, error) {
	err := os.MkdirAll(filepath.Dir(path), dirPermissions)
	if err != nil {
		return nil, fmt.Errorf("logging.Open: create log directory: %w", err)
	}

	err = rotate(path)
	if err != nil {
		return nil, fmt.Errorf("logging.Open: %w", err)
	}

	file, err := os.OpenFile(path, appendFlags, filePermissions) //nolint:gosec // G304: the log path from xdg
	if err != nil {
		return nil, fmt.Errorf("logging.Open: open log file: %w", err)
	}

	return &Log{file: file, logger: newLogger(file)}, nil
}

func (l *Log) Logger() *slog.Logger {
	return l.logger
}

func (l *Log) Close() error {
	err := l.file.Close()
	if err != nil {
		return fmt.Errorf("logging.Log.Close: %w", err)
	}

	return nil
}

func newLogger(file *os.File) *slog.Logger {
	handler := slog.NewTextHandler(file, &slog.HandlerOptions{AddSource: true, Level: slog.LevelInfo, ReplaceAttr: nil})

	return slog.New(handler).With(slog.String(keySession, uuid.NewV7().String()))
}
