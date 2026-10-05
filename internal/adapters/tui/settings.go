package tui

import (
	"time"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
)

type Settings struct {
	Keys          binding.Keys
	ForcedQuitKey string
	Theme         look.Theme
	Location      *time.Location
	Remembered    mainscreen.Remembered
}
