package logging

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
)

const (
	rotateAboveBytes = 5 << 20
	rotatedFilesKept = 2
)

func rotate(path string) error {
	info, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("check log size: %w", err)
	}

	if info.Size() <= rotateAboveBytes {
		return nil
	}

	return shiftRotatedFiles(path)
}

func shiftRotatedFiles(path string) error {
	for generation := rotatedFilesKept; generation > 1; generation-- {
		err := renameIfExists(rotatedName(path, generation-1), rotatedName(path, generation))
		if err != nil {
			return err
		}
	}

	return renameIfExists(path, rotatedName(path, 1))
}

func renameIfExists(from, to string) error {
	err := os.Rename(from, to)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("rotate log file: %w", err)
	}

	return nil
}

func rotatedName(path string, generation int) string {
	return path + "." + strconv.Itoa(generation)
}
