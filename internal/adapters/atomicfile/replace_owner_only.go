package atomicfile

import "fmt"

func ReplaceOwnerOnly(path string, data []byte) error {
	err := renameTempOver(path, data, ownerOnlyPermissions)
	if err != nil {
		return fmt.Errorf("atomicfile.ReplaceOwnerOnly: %w", err)
	}

	return nil
}
