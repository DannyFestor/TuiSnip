package tui_test

import (
	"context"
	"log/slog"
	"reflect"
	"slices"
	"strconv"
	"testing"
	"time"
	"uuid"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	wideWidth          = 120
	wideHeight         = 40
	narrowWidth        = 60
	narrowHeight       = 20
	listLongerThanPane = 45
	filedTitle         = "Table test skeleton"
	filedSnippetID     = "0194c3a0-0000-7000-8000-0000000f11ed"
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
	lister                tui.FolderSnippetsLister
	treeLister            tui.FolderTreeLister
	copier                tui.SnippetCopier
	creator               tui.SnippetCreator
	searcher              tui.SnippetSearcher
	folderCreator         tui.FolderCreator
	folderRenamer         tui.FolderRenamer
	folderDeletePreviewer tui.FolderDeletePreviewer
	folderDeleter         tui.FolderDeleter
	sortOrderSaver        tui.SortOrderSaver
	collapsedFoldersSaver tui.CollapsedFoldersSaver
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
	listingInOrder(lister, folderID, domain.SortOrderTitle, snippets...)
}

func listingInOrder(
	lister *MockFolderSnippetsLister,
	folderID domain.FolderID,
	order domain.SortOrder,
	snippets ...domain.Snippet,
) {
	lister.EXPECT().
		Run(mock.Anything, browse.SnippetsInFolderInput{FolderID: folderID, Order: order}).
		Return(snippets, nil).
		Once()
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
		Lister:        with.lister,
		TreeLister:    with.treeLister,
		Copier:        with.copier,
		Creator:       with.creator,
		Searcher:      with.searcher,
		FolderCreator: orMock(with.folderCreator, func() tui.FolderCreator { return NewMockFolderCreator(t) }),
		FolderRenamer: orMock(with.folderRenamer, func() tui.FolderRenamer { return NewMockFolderRenamer(t) }),
		FolderDeletePreviewer: orMock(
			with.folderDeletePreviewer,
			func() tui.FolderDeletePreviewer { return NewMockFolderDeletePreviewer(t) },
		),
		FolderDeleter:  orMock(with.folderDeleter, func() tui.FolderDeleter { return NewMockFolderDeleter(t) }),
		SortOrderSaver: orMock(with.sortOrderSaver, func() tui.SortOrderSaver { return NewMockSortOrderSaver(t) }),
		CollapsedFoldersSaver: orMock(
			with.collapsedFoldersSaver,
			func() tui.CollapsedFoldersSaver { return NewMockCollapsedFoldersSaver(t) },
		),
		Settings: settings,
		Logger:   slog.New(slog.DiscardHandler),
	})
	require.NoError(t, err)

	return model
}

func orMock[T comparable](given T, newMock func() T) T {
	var unset T
	if given == unset {
		return newMock()
	}

	return given
}

func listerOf(t *testing.T, snippets ...domain.Snippet) *MockFolderSnippetsLister {
	t.Helper()

	lister := NewMockFolderSnippetsLister(t)
	lister.EXPECT().Run(mock.Anything, atRootByTitle()).Return(snippets, nil)

	return lister
}

func listerReturning(t *testing.T, first, then []domain.Snippet) *MockFolderSnippetsLister {
	t.Helper()

	lister := NewMockFolderSnippetsLister(t)
	listingIn(lister, domain.FolderID{}, first...)
	listingIn(lister, domain.FolderID{}, then...)

	return lister
}

func atRootByTitle() browse.SnippetsInFolderInput {
	return browse.SnippetsInFolderInput{FolderID: domain.FolderID{}, Order: domain.SortOrderTitle}
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

func filedIn(t *testing.T, folderID domain.FolderID) domain.Snippet {
	t.Helper()

	return testkit.Snippet(t, testkit.SnippetSpec{
		ID:       domain.SnippetID(uuid.MustParse(filedSnippetID)),
		Title:    filedTitle,
		FolderID: folderID,
		Fragment: testkit.FragmentSpec{
			Language: "Go",
			Content:  "func TestParse(t *testing.T) {}\n",
		},
	})
}

func browsingModel(t *testing.T, sample foldertree.Sample, lister *MockFolderSnippetsLister) tui.Model {
	t.Helper()

	return modelWith(t, browsingActions(t, sample, lister))
}

func browsingActions(t *testing.T, sample foldertree.Sample, lister *MockFolderSnippetsLister) actions {
	t.Helper()

	return actions{
		lister:     lister,
		treeLister: treeOf(t, sample.Tree),
		copier:     NewMockSnippetCopier(t),
		creator:    NewMockSnippetCreator(t),
		searcher:   NewMockSnippetSearcher(t),
	}
}

func hitsOf(snippets ...domain.Snippet) []domain.SearchHit {
	hits := make([]domain.SearchHit, 0, len(snippets))
	for index := range snippets {
		hits = append(hits, domain.NewSearchHit(snippets[index], domain.FieldScores{}))
	}

	return hits
}

func copied(t *testing.T, delivery domain.CopyDelivery) snippet.CopyResult {
	t.Helper()

	content, err := value.NewContent("echo copied")
	require.NoError(t, err)

	return snippet.CopyResult{Delivery: delivery, Content: content}
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

	reflected := reflect.ValueOf(msg)
	if reflected.Kind() != reflect.Slice || !reflected.Type().ConvertibleTo(cmds) {
		return nil, false
	}

	return reflect.TypeAssert[[]tea.Cmd](reflected.Convert(cmds))
}

func (d *driver) screen() string {
	return ansi.Strip(d.model.View().Content)
}
