package library

import (
	"context"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type libraryInsert func(ctx context.Context, library domain.Library) error

// Every Import goes through here, whatever file it read, so each gets the same checks,
// the same all-or-nothing write, and the same counts in its report.
func insertWhole(ctx context.Context, insert libraryInsert, imported domain.Library) (Counts, error) {
	err := imported.RequireIntact()
	if err != nil {
		return Counts{}, fmt.Errorf("check library: %w", err)
	}

	err = insert(ctx, imported)
	if err != nil {
		return Counts{}, err
	}

	return countsOf(imported), nil
}
