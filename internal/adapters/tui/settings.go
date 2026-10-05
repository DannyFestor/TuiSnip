package tui

import (
	"time"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Settings struct {
	Keys          binding.Keys
	ForcedQuitKey string
	Location      *time.Location
	SortOrder     domain.SortOrder
}
