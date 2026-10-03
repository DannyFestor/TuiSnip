package tui

import (
	"time"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
)

type Settings struct {
	Keys          binding.Keys
	ForcedQuitKey string
	Location      *time.Location
}
