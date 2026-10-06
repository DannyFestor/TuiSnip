package snippet

import (
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type UpdateInput struct {
	SnippetID       domain.SnippetID
	LoadedUpdatedAt time.Time
	Title           string
	Description     string
	Language        string
	Content         string
	Tags            []domain.Tag
	NewTags         []string
}

func (in UpdateInput) raw() rawFields {
	return rawFields{
		title:       in.Title,
		description: in.Description,
		language:    in.Language,
		content:     in.Content,
		newTags:     in.NewTags,
	}
}
