package snippetlist

import "github.com/DannyFestor/TuiSnip/internal/domain"

type Meta func(domain.Snippet) string

func Language(snippet domain.Snippet) string {
	return snippet.FirstFragment().Language().String()
}
