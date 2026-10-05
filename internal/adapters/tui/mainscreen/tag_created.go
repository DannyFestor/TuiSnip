package mainscreen

import (
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type TagCreated struct {
	Tags []browse.TagCount
	ID   domain.TagID
}
