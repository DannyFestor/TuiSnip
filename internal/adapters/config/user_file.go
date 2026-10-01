package config

import (
	"errors"
	"fmt"
	"io"
	"os"
)

const maxFileBytes = 1 << 20

func readUserFile(path string) ([]byte, error) {
	file, err := os.Open(path) //nolint:gosec // G304: config.toml from xdg
	if err != nil {
		return nil, fmt.Errorf("open config file: %w", err)
	}

	data, err := io.ReadAll(io.LimitReader(file, maxFileBytes+1))

	err = errors.Join(err, file.Close())
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	if len(data) > maxFileBytes {
		return nil, ErrFileTooLarge
	}

	return data, nil
}
