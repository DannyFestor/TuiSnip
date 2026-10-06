package editor

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	scratchDirPattern = "tuisnip-edit-"
	scratchFileMode   = 0o600
)

type scratchDir struct {
	path string
}

func newScratchDir() (scratchDir, error) {
	path, err := os.MkdirTemp("", scratchDirPattern)
	if err != nil {
		return scratchDir{}, fmt.Errorf("create scratch dir: %w", err)
	}

	return scratchDir{path: path}, nil
}

func (d scratchDir) written(fileName, content string) (string, error) {
	path := filepath.Join(d.path, fileName)

	err := os.WriteFile(path, []byte(content), scratchFileMode)
	if err != nil {
		return "", fmt.Errorf("write scratch file: %w", err)
	}

	return path, nil
}

func (d scratchDir) remove() error {
	err := os.RemoveAll(d.path)
	if err != nil {
		return fmt.Errorf("remove scratch dir: %w", err)
	}

	return nil
}

func readBack(path string) (string, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: the scratch file this package just wrote
	if err != nil {
		return "", fmt.Errorf("read scratch file: %w", err)
	}

	return string(data), nil
}
