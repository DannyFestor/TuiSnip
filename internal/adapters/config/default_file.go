package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	defaultconfig "github.com/DannyFestor/TuiSnip/embeds/config"
)

const (
	directoryPermissions = 0o755
	filePermissions      = 0o644
	tempFileSuffix       = "-*.tmp"
	createNewFlags       = os.O_WRONLY | os.O_CREATE | os.O_EXCL
)

func writeDefaultIfMissing(path string) (bool, error) {
	_, err := os.Lstat(path)
	if err == nil {
		return false, nil
	}

	if !errors.Is(err, fs.ErrNotExist) {
		return false, fmt.Errorf("check config file: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(path), directoryPermissions)
	if err != nil {
		return false, fmt.Errorf("create config directory: %w", err)
	}

	return linkDefaultFile(path)
}

func linkDefaultFile(path string) (bool, error) {
	tempPath, err := writeTempDefault(path)
	if err != nil {
		return false, err
	}

	created, err := linkOrWriteExclusive(tempPath, path)

	return created, errors.Join(err, removeFile(tempPath))
}

func writeTempDefault(path string) (string, error) {
	file, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+tempFileSuffix)
	if err != nil {
		return "", fmt.Errorf("create temporary config file: %w", err)
	}

	err = writeDefaultAndClose(file)
	if err != nil {
		return "", errors.Join(err, removeFile(file.Name()))
	}

	return file.Name(), nil
}

func linkOrWriteExclusive(tempPath, path string) (bool, error) {
	err := os.Link(tempPath, path)
	if errors.Is(err, fs.ErrExist) {
		return false, nil
	}

	if err != nil {
		return writeExclusive(path)
	}

	return true, nil
}

func writeExclusive(path string) (bool, error) {
	file, err := os.OpenFile(path, createNewFlags, filePermissions) //nolint:gosec // G304: config.toml from xdg
	if errors.Is(err, fs.ErrExist) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("create config file: %w", err)
	}

	err = writeDefaultAndClose(file)
	if err != nil {
		return false, errors.Join(err, removeFile(path))
	}

	return true, nil
}

func writeDefaultAndClose(file *os.File) error {
	_, err := file.Write(defaultconfig.Default())

	err = errors.Join(err, file.Chmod(filePermissions), file.Sync(), file.Close())
	if err != nil {
		return fmt.Errorf("write default config: %w", err)
	}

	return nil
}

func removeFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("remove temporary config file: %w", err)
	}

	return nil
}
