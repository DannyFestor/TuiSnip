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
	Content         string
}
