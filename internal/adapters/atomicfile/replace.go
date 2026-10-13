package atomicfile

import (
	"errors"
	"fmt"
	"os"
)

func Replace(path string, data []byte) error {
	err := createDirectoryFor(path)
	if err != nil {
		return fmt.Errorf("atomicfile.Replace: %w", err)
	}

	err = renameTempOver(path, data, filePermissions)
	if err != nil {
		return fmt.Errorf("atomicfile.Replace: %w", err)
	}

	return nil
}

func renameTempOver(path string, data []byte, permissions os.FileMode) error {
	tempPath, err := writeTempBeside(path, data, permissions)
	if err != nil {
		return err
	}

	err = os.Rename(tempPath, path)
	if err != nil {
		return errors.Join(fmt.Errorf("rename over file: %w", err), removeTemp(tempPath))
	}

	return nil
}
