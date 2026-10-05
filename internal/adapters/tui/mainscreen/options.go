package mainscreen

import (
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type Options struct {
	Location   *time.Location
	Remembered Remembered
	Languages  []value.Language
}
