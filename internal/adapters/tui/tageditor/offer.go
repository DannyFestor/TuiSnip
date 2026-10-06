package tageditor

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagchoice"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
)

type Offer struct {
	Listed []browse.TagCount
	Chosen tagchoice.Chosen
}
