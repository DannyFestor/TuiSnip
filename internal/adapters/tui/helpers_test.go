package tui_test

import (
	"log/slog"
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
)

const (
	wideWidth          = 120
	wideHeight         = 40
	narrowWidth        = 60
	narrowHeight       = 20
	minimumWidth       = 80
	minimumHeight      = 24
	listLongerThanPane = 45
)

func defaultSettings() tui.Settings {
	return tui.Settings{
		Global: tui.GlobalKeyMap{
			Quit:         []string{"q"},
			Help:         []string{"?"},
			Search:       []string{"/"},
			NewSnippet:   []string{"n"},
			Capture:      []string{"p"},
			FocusNext:    []string{"tab"},
			FocusPrev:    []string{"shift+tab"},
			FocusRight:   []string{"l", "right"},
			FocusLeft:    []string{"h", "left"},
			FocusFolders: []string{"1"},
			FocusTags:    []string{"2"},
			FocusList:    []string{"3"},
			FocusSnippet: []string{"4"},
			Open:         []string{"enter"},
			Back:         []string{"esc"},
			Down:         []string{"j", "down"},
			Up:           []string{"k", "up"},
			Top:          []string{"g", "home"},
			Bottom:       []string{"G", "end"},
			PageDown:     []string{"pgdown", "ctrl+d"},
			PageUp:       []string{"pgup", "ctrl+u"},
		},
		Folders:     tui.FoldersKeyMap{NewFolder: []string{"N"}},
		SnippetList: tui.SnippetListKeyMap{Copy: []string{"y"}},
		SnippetPane: tui.SnippetPaneKeyMap{Copy: []string{"y"}},
		Editor: tui.EditorKeyMap{
			Save:      []string{"ctrl+s"},
			Cancel:    []string{"esc"},
			NextField: []string{"down", "tab"},
			PrevField: []string{"up", "shift+tab"},
			OpenField: []string{"enter"},
		},
		Content: tui.ContentKeyMap{Save: []string{"ctrl+s"}, Leave: []string{"esc"}},
		Search: tui.SearchKeyMap{
			Down:   []string{"down", "ctrl+n", "ctrl+j"},
			Up:     []string{"up", "ctrl+p", "ctrl+k"},
			Accept: []string{"enter"},
			Copy:   []string{"ctrl+y"},
			Cancel: []string{"esc"},
		},
		Confirm:  tui.ConfirmKeyMap{Yes: []string{"y"}, No: []string{"n", "esc", "enter"}},
		Location: time.UTC,
	}
}

func newModel(t *testing.T, lister tui.FolderSnippetsLister, copier tui.SnippetCopier) tui.Model {
	t.Helper()

	return modelWith(t, actions{
		lister:   lister,
		copier:   copier,
		creator:  NewMockSnippetCreator(t),
		searcher: NewMockSnippetSearcher(t),
	})
}

type actions struct {
	lister   tui.FolderSnippetsLister
	copier   tui.SnippetCopier
	creator  tui.SnippetCreator
	searcher tui.SnippetSearcher
}

func modelWith(t *testing.T, with actions) tui.Model {
	t.Helper()

	model, err := tui.New(t.Context(), tui.Deps{
		Lister:   with.lister,
		Copier:   with.copier,
		Creator:  with.creator,
		Searcher: with.searcher,
		Settings: defaultSettings(),
		Logger:   slog.New(slog.DiscardHandler),
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
		for _, batched := range msg {
			d.run(batched)
		}
	default:
		d.emitted = append(d.emitted, msg)
		d.send(msg)
	}
}

func (d *driver) screen() string {
	return ansi.Strip(d.model.View().Content)
}

func letter(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

func special(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code}
}

func ctrl(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Mod: tea.ModCtrl}
}

func typed(text string) []tea.KeyPressMsg {
	pressed := make([]tea.KeyPressMsg, 0, len(text))
	for _, r := range text {
		pressed = append(pressed, letter(r))
	}

	return pressed
}
