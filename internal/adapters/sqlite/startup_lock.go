package sqlite

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const (
	lockFileSuffix      = ".lock"
	lockFilePermissions = 0o600
	lockTimeout         = 30 * time.Second
	lockPollInterval    = 50 * time.Millisecond
)

type startupLock struct {
	file *os.File
}

func acquireStartupLock(ctx context.Context, databasePath string, logger *slog.Logger) (*startupLock, error) {
	path := filepath.Clean(databasePath + lockFileSuffix)

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, lockFilePermissions)
	if err != nil {
		return nil, fmt.Errorf("open lock file: %w", err)
	}

	ctx, cancel := context.WithTimeoutCause(ctx, lockTimeout, ErrLockTimeout)
	defer cancel()

	err = waitForLock(ctx, file, logger)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("lock %s: %w", path, err), file.Close())
	}

	return &startupLock{file: file}, nil
}

func waitForLock(ctx context.Context, file *os.File, logger *slog.Logger) error {
	locked, err := tryLock(file)
	if locked || err != nil {
		return err
	}

	logger.DebugContext(ctx, "waiting for the start-up lock", slog.String(keyLockPath, file.Name()))

	ticker := time.NewTicker(lockPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("give up waiting: %w", context.Cause(ctx))
		case <-ticker.C:
			locked, err = tryLock(file)
			if locked || err != nil {
				return err
			}
		}
	}
}

func tryLock(file *os.File) (bool, error) {
	err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("flock: %w", err)
	}

	return true, nil
}

func (l *startupLock) release() error {
	err := errors.Join(syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN), l.file.Close())
	if err != nil {
		return fmt.Errorf("release start-up lock: %w", err)
	}

	return nil
}
