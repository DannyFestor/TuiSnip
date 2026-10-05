package tui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func TestModel_collapsedFolders(t *testing.T) {
	t.Parallel()

	t.Run("starts with the remembered Folders collapsed", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		settings := testsettings.Default(t)
		settings.Remembered.CollapsedFolders = []domain.FolderID{sample.Go.ID()}

		screen := start(
			t,
			modelWithSettings(t, collapsing(t, sample, listerOf(t), nil), settings),
			wideWidth,
			wideHeight,
		)

		assert.Regexp(t, `▸ go +2`, screen.screen())
		assert.NotContains(t, screen.screen(), "testing")
	})

	t.Run("forgets a collapsed id no Folder has once the tree loads", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		saver := NewMockCollapsedFoldersSaver(t)
		saver.EXPECT().SaveCollapsedFolders(mock.Anything, []domain.FolderID{sample.Go.ID()}).Return(nil).Once()

		settings := testsettings.Default(t)
		settings.Remembered.CollapsedFolders = []domain.FolderID{
			testkit.NewSequentialIDs().NewFolderID(),
			sample.Go.ID(),
		}

		screen := start(
			t,
			modelWithSettings(t, collapsing(t, sample, listerOf(t), saver), settings),
			wideWidth,
			wideHeight,
		)

		assert.Regexp(t, `▸ go +2`, screen.screen())
	})

	t.Run("space remembers the collapsed Folder and hides its subtree", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		saver := NewMockCollapsedFoldersSaver(t)
		saver.EXPECT().SaveCollapsedFolders(mock.Anything, []domain.FolderID{sample.Go.ID()}).Return(nil).Once()
		screen := start(t, modelWith(t, collapsing(t, sample, listerThrough(t, sample), saver)), wideWidth, wideHeight)

		screen.press(keypress.Typed("jj ")...)

		assert.Regexp(t, `▸ go +2`, screen.screen())
		assert.NotContains(t, screen.screen(), "testing")
	})

	t.Run("a failed save shows the failure and keeps the Folder collapsed", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		saver := NewMockCollapsedFoldersSaver(t)
		saver.EXPECT().SaveCollapsedFolders(mock.Anything, mock.Anything).Return(errDatabaseLocked).Once()
		screen := start(t, modelWith(t, collapsing(t, sample, listerThrough(t, sample), saver)), wideWidth, wideHeight)

		screen.press(keypress.Typed("jj ")...)

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
		assert.Regexp(t, `▸ go +2`, screen.screen())
	})
}

func collapsing(
	t *testing.T,
	sample foldertree.Sample,
	lister *MockFolderSnippetsLister,
	saver *MockCollapsedFoldersSaver,
) actions {
	t.Helper()

	with := browsingActions(t, sample, lister)
	if saver != nil {
		with.collapsedFoldersSaver = saver
	}

	return with
}

func listerThrough(t *testing.T, sample foldertree.Sample) *MockFolderSnippetsLister {
	t.Helper()

	lister := listerOf(t)
	listingIn(lister, sample.Docker.ID())
	listingIn(lister, sample.Go.ID())

	return lister
}
