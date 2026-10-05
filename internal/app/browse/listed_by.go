package browse

import (
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type listing func(snippets []domain.Snippet, err error) ([]domain.Snippet, error)

func listedBy(action string) listing {
	return func(snippets []domain.Snippet, err error) ([]domain.Snippet, error) {
		if err != nil {
			return nil, fmt.Errorf("%s: %w", action, err)
		}

		return snippets, nil
	}
}
