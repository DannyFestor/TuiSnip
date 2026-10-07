package tui

import (
	"time"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type Settings struct {
	Keys          binding.Keys
	ForcedQuitKey string
	Theme         look.Theme
	Mouse         bool
	Languages     []value.Language
	Location      *time.Location
	Remembered    mainscreen.Remembered
}
