package editoverlay

import (
	"slices"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type browsedFragment struct {
	browsed  BrowsedSnippet
	fragment domain.Fragment
}

func (b browsedFragment) shown() int {
	return max(0, slices.IndexFunc(b.browsed.Snippet.Fragments(), func(fragment domain.Fragment) bool {
		return fragment.ID() == b.fragment.ID()
	}))
}
