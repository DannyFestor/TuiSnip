package state

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

func replaceAtomically(path string, data []byte) error {
	err := os.MkdirAll(filepath.Dir(path), directoryPermissions)
	if err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}

	tempPath, err := writeTemp(path, data)
	if err != nil {
		return err
	}

	err = os.Rename(tempPath, path)
	if err != nil {
		return errors.Join(fmt.Errorf("replace state file: %w", err), removeFile(tempPath))
	}

	return nil
}

func writeTemp(path string, data []byte) (string, error) {
	file, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+tempFileSuffix)
	if err != nil {
		return "", fmt.Errorf("create temporary state file: %w", err)
	}

	_, err = file.Write(data)

	err = errors.Join(err, file.Chmod(filePermissions), file.Sync(), file.Close())
	if err != nil {
		return "", errors.Join(fmt.Errorf("write temporary state file: %w", err), removeFile(file.Name()))
	}

	return file.Name(), nil
}

func removeFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("remove temporary state file: %w", err)
	}

	return nil
}
