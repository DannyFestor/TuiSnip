package tui

import (
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	goneElsewhereFormat   = "%s %q was moved to the Trash or deleted in another TuiSnip"
	noLongerInTrashFormat = "%q is no longer in the Trash"
)

type gone struct {
	notFound   string
	folderGone string
}

func goneElsewhere(noun, name string) gone {
	return gone{notFound: fmt.Sprintf(goneElsewhereFormat, noun, name), folderGone: ""}
}

func goneFromTrash(name string) gone {
	return gone{notFound: fmt.Sprintf(noLongerInTrashFormat, name), folderGone: ""}
}

func (g gone) placingInto(folderName string) gone {
	g.folderGone = fmt.Sprintf(goneElsewhereFormat, folderNoun, folderName)

	return g
}

func (g gone) text(err error) (string, bool) {
	switch {
	case errors.Is(err, domain.ErrFolderGone) && g.folderGone != "":
		return g.folderGone, true
	case errors.Is(err, domain.ErrNotFound), errors.Is(err, domain.ErrFolderGone):
		return g.notFound, g.notFound != ""
	}

	return "", false
}
