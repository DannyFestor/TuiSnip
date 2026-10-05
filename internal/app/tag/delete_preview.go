package tag

import "github.com/DannyFestor/TuiSnip/internal/domain"

type DeletePreview struct {
	Tag          domain.Tag
	SnippetCount int
}
