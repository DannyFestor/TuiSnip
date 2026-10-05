package domain

import "fmt"

//go:generate go-enum --marshal

// ENUM(title, updated, created).
type SortOrder string

func (o SortOrder) Next() (SortOrder, error) {
	switch o {
	case SortOrderTitle:
		return SortOrderUpdated, nil
	case SortOrderUpdated:
		return SortOrderCreated, nil
	case SortOrderCreated:
		return SortOrderTitle, nil
	}

	return "", fmt.Errorf("domain.SortOrder.Next: %q: %w", o, ErrInvalidSortOrder)
}
