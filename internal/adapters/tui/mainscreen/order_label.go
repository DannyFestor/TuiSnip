package mainscreen

import "github.com/DannyFestor/TuiSnip/internal/domain"

const (
	byTitleLabel        = "by title"
	byLastUpdatedLabel  = "by last updated"
	byCreationDateLabel = "by creation date"
)

func orderLabel(order domain.SortOrder) string {
	switch order {
	case domain.SortOrderTitle:
		return byTitleLabel
	case domain.SortOrderUpdated:
		return byLastUpdatedLabel
	case domain.SortOrderCreated:
		return byCreationDateLabel
	}

	return ""
}
