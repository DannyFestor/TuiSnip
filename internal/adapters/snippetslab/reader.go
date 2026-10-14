package snippetslab

import (
	"context"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Reader struct {
	ids   IDGenerator
	clock Clock
}

func NewReader(ids IDGenerator, clock Clock) *Reader {
	return &Reader{ids: ids, clock: clock}
}

func (r *Reader) Read(_ context.Context, path string) (domain.Library, domain.ImportLosses, error) {
	read, err := contentsAt(path)
	if err != nil {
		return domain.Library{}, domain.ImportLosses{}, fmt.Errorf("snippetslab.Reader.Read: %w", err)
	}

	converted := newConversion(r.ids, r.clock.Now())

	library, err := converted.library(read)
	if err != nil {
		return domain.Library{}, domain.ImportLosses{}, fmt.Errorf("snippetslab.Reader.Read: %w", err)
	}

	return library, converted.losses, nil
}
