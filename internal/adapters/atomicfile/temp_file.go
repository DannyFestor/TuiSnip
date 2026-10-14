package atomicfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	directoryPermissions = 0o755
	filePermissions      = 0o644
	ownerOnlyPermissions = 0o600
	tempFileSuffix       = "-*.tmp"
)

func createDirectoryFor(path string) error {
	err := os.MkdirAll(filepath.Dir(path), directoryPermissions)
	if err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	return nil
}

func writeTempBeside(path string, data []byte, permissions os.FileMode) (string, error) {
	file, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+tempFileSuffix)
	if err != nil {
		return "", fmt.Errorf("create temporary file: %w", err)
	}

	err = writeAndClose(file, data, permissions)
	if err != nil {
		return "", errors.Join(err, removeTemp(file.Name()))
	}

	return file.Name(), nil
}

func writeAndClose(file *os.File, data []byte, permissions os.FileMode) error {
	_, err := file.Write(data)

	err = errors.Join(err, file.Chmod(permissions), file.Sync(), file.Close())
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

func removeTemp(path string) error {
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("remove temporary file: %w", err)
	}

	return nil
}
