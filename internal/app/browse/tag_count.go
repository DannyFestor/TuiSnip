package browse

import "github.com/DannyFestor/TuiSnip/internal/domain"

type TagCount struct {
	Tag          domain.Tag
	SnippetCount int
}
