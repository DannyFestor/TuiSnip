package testapp

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite"
	"github.com/DannyFestor/TuiSnip/internal/adapters/system"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

// InsertSnippet bypasses the Actions because none creates a Snippet with several Fragments yet.
func (h *Home) InsertSnippet(t *testing.T, app *bootstrap.App, snippet domain.Snippet) domain.Snippet {
	t.Helper()

	database, err := sqlite.Open(t.Context(), sqlite.Options{
		Path:   h.Paths.DatabaseFile,
		Logger: slog.New(slog.DiscardHandler),
		Clock:  system.NewClock(),
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, database.Close()) })

	repository := sqlite.NewSnippetRepository(database, slog.New(slog.DiscardHandler))
	require.NoError(t, repository.Insert(t.Context(), snippet))

	return StoredSnippet(t, app, snippet.ID())
}
