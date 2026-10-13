package snippetlist

import (
	"strconv"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const otherFragmentsMarker = " +"

type Meta func(domain.Snippet) string

func Language(snippet domain.Snippet) string {
	language := snippet.FirstFragment().Language().String()

	others := len(snippet.Fragments()) - 1
	if others == 0 {
		return language
	}

	return language + otherFragmentsMarker + strconv.Itoa(others)
}
