package browse

import "github.com/DannyFestor/TuiSnip/internal/domain"

type SnippetsWithTagInput struct {
	TagID domain.TagID
	Order domain.SortOrder
}
