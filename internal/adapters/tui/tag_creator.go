package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type TagCreator interface {
	Run(ctx context.Context, in tag.CreateInput) (domain.Tag, error)
}
