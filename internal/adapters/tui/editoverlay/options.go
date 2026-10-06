package editoverlay

import (
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type Options struct {
	Languages []value.Language
	Tags      []browse.TagCount
}
