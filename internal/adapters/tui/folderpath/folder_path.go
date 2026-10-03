package folderpath

import "github.com/DannyFestor/TuiSnip/internal/domain"

const Root = "Root"

func Of(domain.Snippet) string {
	return Root
}
