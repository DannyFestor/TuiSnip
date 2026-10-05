package mainscreen

import (
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	byTitleLabel        = "by title"
	byLastUpdatedLabel  = "by last updated"
	byCreationDateLabel = "by creation date"
)

func orderLabel(order domain.SortOrder) (string, error) {
	switch order {
	case domain.SortOrderTitle:
		return byTitleLabel, nil
	case domain.SortOrderUpdated:
		return byLastUpdatedLabel, nil
	case domain.SortOrderCreated:
		return byCreationDateLabel, nil
	}

	return "", fmt.Errorf("label sort order %q: %w", order, domain.ErrInvalidSortOrder)
}
