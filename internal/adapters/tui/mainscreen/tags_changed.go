package mainscreen

import (
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type TagsChanged struct {
	Tags      []browse.TagCount
	Selecting domain.TagID
}
