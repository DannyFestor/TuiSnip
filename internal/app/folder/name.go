package folder

import (
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

func parseName(raw string) (value.FolderName, error) {
	name, err := value.NewFolderName(raw)
	if err != nil {
		return value.FolderName{}, fmt.Errorf("parse name: %w", domain.OnField(domain.FieldFolderName, err))
	}

	return name, nil
}
