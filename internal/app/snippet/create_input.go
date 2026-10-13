package snippet

import "github.com/DannyFestor/TuiSnip/internal/domain"

type CreateInput struct {
	Title       string
	Description string
	Fragments   []FragmentInput
	FolderID    domain.FolderID
	Tags        []domain.Tag
	NewTags     []string
}

func (in CreateInput) raw() rawFields {
	return rawFields{
		title:       in.Title,
		description: in.Description,
		fragments:   in.Fragments,
		newTags:     in.NewTags,
	}
}
