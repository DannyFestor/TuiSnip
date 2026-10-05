package tui

import (
	"time"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
)

type Settings struct {
	Keys          binding.Keys
	ForcedQuitKey string
	Location      *time.Location
	Remembered    mainscreen.Remembered
}
