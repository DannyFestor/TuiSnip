package tui

import (
	"errors"
	"fmt"
	"log/slog"

	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const goneElsewhereFormat = "%s %q was moved to the Trash or deleted in another TuiSnip"

type gone struct {
	notFound   string
	folderGone string
}

type itemActionFailedMsg struct {
	operation string
	err       error
	gone      gone
}

func goneItem(noun, name string) gone {
	return gone{notFound: fmt.Sprintf(goneElsewhereFormat, noun, name), folderGone: ""}
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

func (m Model) itemActionFailed(msg itemActionFailedMsg) (Model, tea.Cmd) {
	text, isGone := msg.gone.text(msg.err)
	if !isGone {
		return m.failed(msg.operation, msg.err)
	}

	m.logger.InfoContext(m.ctx, "item gone elsewhere",
		slog.String(keyOperation, msg.operation), slog.Any(keyError, msg.err))

	return m.overlaysUpdated(mainscreen.GoneElsewhereShown{Text: text})
}
