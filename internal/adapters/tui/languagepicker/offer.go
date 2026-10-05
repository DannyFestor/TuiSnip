package languagepicker

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type Offer struct {
	Title   string
	Curated []value.Language
	Current value.Language
	Picked  func(picked value.Language) outcome.Outcome
}
