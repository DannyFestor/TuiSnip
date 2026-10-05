package tag

import (
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

func parseName(raw string) (value.TagName, error) {
	name, err := value.NewTagName(raw)
	if err != nil {
		return value.TagName{}, fmt.Errorf("parse name: %w", domain.OnField(domain.FieldTagName, err))
	}

	return name, nil
}
