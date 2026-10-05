package atomicfile

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

const createNewFlags = os.O_WRONLY | os.O_CREATE | os.O_EXCL

func CreateIfMissing(path string, data []byte) (bool, error) {
	created, err := createUnlessPresent(path, data)
	if err != nil {
		return false, fmt.Errorf("atomicfile.CreateIfMissing: %w", err)
	}

	return created, nil
}

func createUnlessPresent(path string, data []byte) (bool, error) {
	_, err := os.Lstat(path)
	if err == nil {
		return false, nil
	}

	if !errors.Is(err, fs.ErrNotExist) {
		return false, fmt.Errorf("check file: %w", err)
	}

	err = createDirectoryFor(path)
	if err != nil {
		return false, err
	}

	return linkTemp(path, data)
}

func linkTemp(path string, data []byte) (bool, error) {
	tempPath, err := writeTempBeside(path, data)
	if err != nil {
		return false, err
	}

	created, err := linkOrWriteExclusive(tempPath, path, data)

	return created, errors.Join(err, removeTemp(tempPath))
}

func linkOrWriteExclusive(tempPath, path string, data []byte) (bool, error) {
	err := os.Link(tempPath, path)
	if errors.Is(err, fs.ErrExist) {
		return false, nil
	}

	if err != nil {
		return writeExclusive(path, data)
	}

	return true, nil
}

func writeExclusive(path string, data []byte) (bool, error) {
	file, err := os.OpenFile(path, createNewFlags, filePermissions) //nolint:gosec // G304: a path its caller resolved
	if errors.Is(err, fs.ErrExist) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("create file: %w", err)
	}

	err = writeAndClose(file, data)
	if err != nil {
		return false, errors.Join(err, os.Remove(path))
	}

	return true, nil
}
