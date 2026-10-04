package tui_test

import (
	"context"
	"log/slog"
	"reflect"
	"slices"
	"strconv"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	wideWidth          = 120
	wideHeight         = 40
	narrowWidth        = 60
	narrowHeight       = 20
	listLongerThanPane = 45
)

func newModel(t *testing.T, lister tui.FolderSnippetsLister, copier tui.SnippetCopier) tui.Model {
	t.Helper()

	return modelWith(t, actions{
		lister:     lister,
		treeLister: treeOf(t, emptyTree()),
		copier:     copier,
		creator:    NewMockSnippetCreator(t),
		searcher:   NewMockSnippetSearcher(t),
	})
}

type actions struct {
	lister     tui.FolderSnippetsLister
	treeLister tui.FolderTreeLister
	copier     tui.SnippetCopier
	creator    tui.SnippetCreator
	searcher   tui.SnippetSearcher
}

func emptyTree() browse.Tree {
	return browse.Tree{RootSnippetCount: 0, Folders: nil}
}

func treeOf(t *testing.T, tree browse.Tree) *MockFolderTreeLister {
	t.Helper()

	lister := NewMockFolderTreeLister(t)
	lister.EXPECT().Run(mock.Anything, browse.FolderTreeInput{}).Return(tree, nil)

	return lister
}

func listingIn(lister *MockFolderSnippetsLister, folderID domain.FolderID, snippets ...domain.Snippet) {
	lister.EXPECT().Run(mock.Anything, browse.SnippetsInFolderInput{FolderID: folderID}).Return(snippets, nil).Once()
}

func modelWith(t *testing.T, with actions) tui.Model {
	t.Helper()

	return modelWithSettings(t, with, testsettings.Default(t))
}

func modelWithSettings(t *testing.T, with actions, settings tui.Settings) tui.Model {
	t.Helper()

	return modelBuiltBy(t, tui.New, with, settings)
}

type modelConstructor func(context.Context, tui.Deps) (tui.Model, error)

func modelBuiltBy(t *testing.T, build modelConstructor, with actions, settings tui.Settings) tui.Model {
	t.Helper()

	model, err := build(t.Context(), tui.Deps{
		Lister:     with.lister,
		TreeLister: with.treeLister,
		Copier:     with.copier,
		Creator:    with.creator,
		Searcher:   with.searcher,
		Settings:   settings,
		Logger:     slog.New(slog.DiscardHandler),
	})
	require.NoError(t, err)

	return model
}

func listerOf(t *testing.T, snippets ...domain.Snippet) *MockFolderSnippetsLister {
	t.Helper()

	lister := NewMockFolderSnippetsLister(t)
	lister.EXPECT().Run(mock.Anything, browse.SnippetsInFolderInput{FolderID: domain.FolderID{}}).Return(snippets, nil)

	return lister
}

func listerReturning(t *testing.T, first, then []domain.Snippet) *MockFolderSnippetsLister {
	t.Helper()

	lister := NewMockFolderSnippetsLister(t)
	lister.EXPECT().
		Run(mock.Anything, browse.SnippetsInFolderInput{FolderID: domain.FolderID{}}).
		Return(first, nil).
		Once()
	lister.EXPECT().
		Run(mock.Anything, browse.SnippetsInFolderInput{FolderID: domain.FolderID{}}).
		Return(then, nil).
		Once()

	return lister
}

func sampleSnippets(t *testing.T) []domain.Snippet {
	t.Helper()

	ids := testkit.NewSequentialIDs()
	created := time.Date(2026, time.September, 6, 9, 0, 0, 0, time.UTC)

	return []domain.Snippet{
		testkit.Snippet(t, testkit.SnippetSpec{
			ID:          ids.NewSnippetID(),
			Title:       "Graceful HTTP shutdown",
			Description: "Stop accepting, drain, exit",
			Fragment: testkit.FragmentSpec{
				ID:       ids.NewFragmentID(),
				Language: "Go",
				Content:  "func run(ctx context.Context) error {\n\treturn srv.Shutdown(ctx)\n}\n",
			},
			CreatedAt: created,
			UpdatedAt: created.AddDate(0, 0, 21),
		}),
		testkit.Snippet(t, testkit.SnippetSpec{
			ID:          ids.NewSnippetID(),
			Title:       "Prune everything",
			Description: "Reclaim disk space",
			Fragment: testkit.FragmentSpec{
				ID:       ids.NewFragmentID(),
				Language: "Bash",
				Content:  "docker system prune --all --volumes --force\n",
			},
			CreatedAt: created,
		}),
	}
}

func numberedSnippets(t *testing.T, count int) []domain.Snippet {
	t.Helper()

	ids := testkit.NewSequentialIDs()
	snippets := make([]domain.Snippet, 0, count)

	for number := 1; number <= count; number++ {
		snippets = append(snippets, testkit.Snippet(t, testkit.SnippetSpec{
			ID:          ids.NewSnippetID(),
			Title:       "Snippet " + strconv.Itoa(number),
			Description: "Description " + strconv.Itoa(number),
			Fragment:    testkit.FragmentSpec{ID: ids.NewFragmentID()},
		}))
	}

	return snippets
}

type driver struct {
	t       *testing.T
	model   tea.Model
	emitted []tea.Msg
}

func start(t *testing.T, model tui.Model, width, height int) *driver {
	t.Helper()

	screen := &driver{t: t, model: model, emitted: nil}
	screen.send(tea.WindowSizeMsg{Width: width, Height: height})
	screen.run(model.Init())

	return screen
}

func (d *driver) send(msg tea.Msg) {
	d.t.Helper()

	model, cmd := d.model.Update(msg)
	d.model = model
	d.run(cmd)
}

func (d *driver) press(keys ...tea.KeyPressMsg) {
	d.t.Helper()

	for _, pressed := range keys {
		d.send(pressed)
	}
}

func (d *driver) run(cmd tea.Cmd) {
	d.t.Helper()

	if cmd == nil {
		return
	}

	switch msg := cmd().(type) {
	case nil:
	case tea.BatchMsg:
		// Batch promises no order, so running it backwards makes a test that needs an order fail unless the model used Sequence.
		for _, batched := range slices.Backward(msg) {
			d.run(batched)
		}
	default:
		d.runSequenceOrSend(msg)
	}
}

func (d *driver) runSequenceOrSend(msg tea.Msg) {
	d.t.Helper()

	if sequence, ok := sequenceSteps(msg); ok {
		for _, step := range sequence {
			d.run(step)
		}

		return
	}

	d.emitted = append(d.emitted, msg)
	d.send(msg)
}

// tea.Sequence wraps its commands in an unexported slice type, so only its shape identifies it.
func sequenceSteps(msg tea.Msg) ([]tea.Cmd, bool) {
	cmds := reflect.TypeFor[[]tea.Cmd]()

	value := reflect.ValueOf(msg)
	if value.Kind() != reflect.Slice || !value.Type().ConvertibleTo(cmds) {
		return nil, false
	}

	return reflect.TypeAssert[[]tea.Cmd](value.Convert(cmds))
}

func (d *driver) screen() string {
	return ansi.Strip(d.model.View().Content)
}
