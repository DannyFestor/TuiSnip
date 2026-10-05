package domain

//go:generate go-enum --marshal

// ENUM(title, updated, created).
type SortOrder string

func (o SortOrder) Next() SortOrder {
	switch o {
	case SortOrderTitle:
		return SortOrderUpdated
	case SortOrderUpdated:
		return SortOrderCreated
	case SortOrderCreated:
		return SortOrderTitle
	}

	return SortOrderTitle
}
