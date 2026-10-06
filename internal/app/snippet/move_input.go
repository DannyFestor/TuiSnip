package snippet

import "github.com/DannyFestor/TuiSnip/internal/domain"

type MoveInput struct {
	SnippetID domain.SnippetID
	FolderID  domain.FolderID
}
