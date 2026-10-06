package snippet

import "github.com/DannyFestor/TuiSnip/internal/domain"

type IDGenerator interface {
	NewSnippetID() domain.SnippetID
	NewFragmentID() domain.FragmentID
	NewTagID() domain.TagID
}
