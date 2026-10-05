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
	tempFileSuffix       = "-*.tmp"
)

func createDirectoryFor(path string) error {
	err := os.MkdirAll(filepath.Dir(path), directoryPermissions)
	if err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	return nil
}

func writeTempBeside(path string, data []byte) (string, error) {
	file, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+tempFileSuffix)
	if err != nil {
		return "", fmt.Errorf("create temporary file: %w", err)
	}

	err = writeAndClose(file, data)
	if err != nil {
		return "", errors.Join(err, removeTemp(file.Name()))
	}

	return file.Name(), nil
}

func writeAndClose(file *os.File, data []byte) error {
	_, err := file.Write(data)

	err = errors.Join(err, file.Chmod(filePermissions), file.Sync(), file.Close())
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
