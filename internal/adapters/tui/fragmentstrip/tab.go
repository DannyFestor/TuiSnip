package fragmentstrip

import (
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type Tab struct {
	Title    value.FragmentTitle
	Language value.Language
}

func tabOf(fragment domain.Fragment) Tab {
	return Tab{Title: fragment.Title(), Language: fragment.Language()}
}

func (t Tab) Label() string {
	if t.Title.IsSet() {
		return t.Title.String()
	}

	return t.Language.String()
}
