package mainscreen

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/confirm"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/editoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/helpoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/searchpopup"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Screen struct {
	keys            binding.Keys
	global          binding.Set
	listKeys        binding.Set
	paneKeys        binding.Set
	styles          look.Styles
	box             look.Size
	layout          layout
	focus           pane
	selectionHolder pane
	zoomed          bool
	panes           panes
	status          string
}

func New(keys binding.Keys, styles look.Styles, location *time.Location, remembered Remembered) (Screen, error) {
	label, err := orderLabel(remembered.SortOrder)
	if err != nil {
		return Screen{}, fmt.Errorf("mainscreen.New: %w", err)
	}

	return Screen{
		keys:            keys,
		global:          keys.For(binding.ScopeGlobal),
		listKeys:        keys.For(binding.ScopeSnippetList),
		paneKeys:        keys.For(binding.ScopeSnippetPane),
		styles:          styles,
		box:             look.Size{Width: 0, Height: 0},
		layout:          arrange(look.Size{Width: 0, Height: 0}, paneFolders, paneFolders),
		focus:           paneFolders,
		selectionHolder: paneFolders,
		zoomed:          false,
		panes:           newPanes(keys, styles, location, remembered.CollapsedFolders).withOrderLabel(label),
		status:          "",
	}, nil
}

func (s Screen) Update(msg tea.Msg) outcome.Step {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return s.pressed(msg)
	case tea.PasteMsg:
		return s.pasted(msg)
	case look.Resized:
		return outcome.Stay(s.resized(msg.Box))
	case look.Restyled:
		return outcome.Stay(s.restyled(msg))
	case FolderDeletePreviewed:
		return s.confirmingFolderDelete(msg.Preview)
	case StatusShown:
		return outcome.Stay(s.withStatus(msg.Text))
	case Captured:
		return s.opening(editoverlay.Capturing(s.keys, s.styles, s.destination(), msg.Content))
	}

	return s.loaded(msg)
}

func (s Screen) Received(received outcome.Outcome) outcome.Step {
	switch received := received.(type) {
	case outcome.SnippetRevealed:
		next, expanded := s.revealing(received.FolderID)

		return outcome.Stay(next).Passing(expanded...).Passing(received)
	default:
		return outcome.Stay(s).Passing(received)
	}
}

func (s Screen) View() string {
	return s.ViewUnder(s.ShortHelp())
}

func (s Screen) ViewUnder(hints []key.Binding) string {
	return s.panesView() + "\n" + statusLine(s.styles, s.status, s.hint(hints), s.box.Width)
}

func (s Screen) ShortHelp() []key.Binding {
	if tooSmall(s.box) {
		return nil
	}

	return s.panes.keyMaps()[s.focus].ShortHelp()
}

func (s Screen) FullHelp() [][]key.Binding {
	return append(s.global.FullHelp(), s.panes.keyMaps()[s.focus].FullHelp()...)
}

func (s Screen) loaded(msg tea.Msg) outcome.Step {
	switch msg := msg.(type) {
	case TreeLoaded:
		return s.treeLoaded(msg.Tree)
	case TreeChanged:
		return s.treeChanged(msg)
	case TagsLoaded:
		return outcome.Stay(s.withPanes(s.panes.withTags(msg.Tags)))
	case SnippetsLoaded:
		return s.snippetsLoaded(msg)
	}

	return outcome.Stay(s)
}

func (s Screen) pressed(msg tea.KeyPressMsg) outcome.Step {
	if s.panes.naming() {
		return s.focusedUpdated(msg)
	}

	if step, ok := s.globalPressed(msg); ok {
		return step
	}

	if s.focus == paneList && s.listKeys.Matches(msg, binding.CycleSort) {
		return outcome.Stay(s).Passing(s.panes.sortCycleAsked(s.selection()))
	}

	if stored, ok := s.editAsked(msg); ok {
		return s.opening(editoverlay.Editing(
			s.keys,
			s.styles,
			editoverlay.BrowsedSnippet{Snippet: stored, Selection: s.selection()},
		))
	}

	return s.focusedUpdated(msg)
}

func (s Screen) globalPressed(msg tea.KeyPressMsg) (outcome.Step, bool) {
	switch {
	case s.global.Matches(msg, binding.Quit):
		return outcome.Stay(s).Passing(outcome.QuitAsked{}), true
	case s.global.Matches(msg, binding.Help):
		return outcome.Stay(s).Opening(helpoverlay.New(s.keys, s.styles, s.FullHelp())), true
	case s.global.Matches(msg, binding.NewSnippet):
		return s.opening(editoverlay.New(s.keys, s.styles, s.destination())), true
	case s.global.Matches(msg, binding.Capture):
		return outcome.Stay(s).Passing(outcome.CaptureAsked{}), true
	case s.global.Matches(msg, binding.Search):
		return s.opening(searchpopup.New(s.keys, s.styles, s.panes.preview.Cleared(), s.panes.listing())), true
	case s.global.Matches(msg, binding.Zoom):
		return outcome.Stay(s.zoomToggled()), true
	case s.focus.inLeftColumn() && s.global.Matches(msg, binding.Open):
		return s.opened(), true
	}

	if navigate, ok := navigationFor(s.global, msg); ok {
		return outcome.Stay(s.focusedOn(navigate(s.focus, s.selectionHolder))), true
	}

	return outcome.Stay(s), false
}

func (s Screen) editAsked(msg tea.KeyPressMsg) (domain.Snippet, bool) {
	switch {
	case s.focus == paneList && s.listKeys.Matches(msg, binding.Edit):
		return s.panes.list.Selected()
	case s.focus == paneSnippet && s.paneKeys.Matches(msg, binding.Edit):
		return s.panes.preview.Shown()
	}

	return domain.Snippet{}, false
}

func (s Screen) opened() outcome.Step {
	next := s.holding(s.focus)
	step := outcome.Stay(next.focusedOn(s.focus.drillIn(s.selectionHolder)))

	if next.selection() == s.selection() {
		return step
	}

	return step.Passing(selected(next.selection())...)
}

func (s Screen) holding(holder pane) Screen {
	if _, ok := s.panes.tags.Selected(); holder == paneTags && !ok {
		return s
	}

	s.selectionHolder = holder

	return s
}

func (s Screen) selection() browseselection.Selection {
	return s.panes.selectionIn(s.selectionHolder)
}

func (s Screen) destination() editoverlay.Destination {
	return s.panes.destination(s.selection())
}

func selected(selection browseselection.Selection) []outcome.Outcome {
	if tagID, ok := selection.Tag(); ok {
		return []outcome.Outcome{outcome.TagSelected{ID: tagID}}
	}

	folderID, _ := selection.Folder()

	return []outcome.Outcome{outcome.FolderSelected{ID: folderID}}
}

func (s Screen) pasted(msg tea.PasteMsg) outcome.Step {
	if !s.panes.naming() {
		return outcome.Stay(s)
	}

	return s.focusedUpdated(msg)
}

func (s Screen) treeLoaded(tree browse.Tree) outcome.Step {
	next, pruned := s.panes.withTree(tree)

	return outcome.Stay(s.withPanes(next)).Passing(pruned...)
}

func (s Screen) treeChanged(msg TreeChanged) outcome.Step {
	withTree, pruned := s.panes.withTree(msg.Tree)
	next, expanded := s.withPanes(withTree).selectingFolder(msg.Selecting)

	return outcome.Stay(next).
		Passing(pruned...).
		Passing(expanded...).
		Passing(outcome.FolderSelected{ID: msg.Selecting})
}

func (s Screen) snippetsLoaded(loaded SnippetsLoaded) outcome.Step {
	next, err := s.panes.withSnippetsIfStillSelected(loaded, s.selection())
	step := outcome.Stay(s.withPanes(next))

	if err != nil {
		return step.Passing(outcome.SortOrderRejected{Err: err})
	}

	return step
}

func (s Screen) confirmingFolderDelete(preview folder.DeletePreview) outcome.Step {
	onYes := outcome.FolderDeleteRequested{
		Input:    folder.DeleteInput{FolderID: preview.Folder.ID()},
		ParentID: preview.Folder.ParentID(),
	}

	return outcome.Stay(s).Opening(confirm.New(s.keys, s.styles, folderDeleteQuestion(preview), onYes))
}

func (s Screen) opening(child outcome.Overlay, cmd tea.Cmd) outcome.Step {
	return outcome.Stay(s).Opening(child).Running(cmd)
}

func (s Screen) focusedUpdated(msg tea.Msg) outcome.Step {
	updated, outcomes, cmd := s.panes.updated(s.focus, msg)
	next := s.withPanes(updated).holdingAfter(outcomes).arranged()

	return outcome.Stay(next).Passing(outcomes...).Running(cmd)
}

func (s Screen) holdingAfter(outcomes []outcome.Outcome) Screen {
	for _, reported := range outcomes {
		switch reported.(type) {
		case outcome.FolderSelected:
			s.selectionHolder = paneFolders
		case outcome.TagSelected:
			s.selectionHolder = paneTags
		default:
		}
	}

	return s
}

func (s Screen) withPanes(next panes) Screen {
	s.panes = next

	return s
}

func (s Screen) restyled(msg look.Restyled) Screen {
	s.styles = msg.Styles
	s.panes = s.panes.restyled(msg)

	return s
}

func (s Screen) withStatus(text string) Screen {
	s.status = text

	return s
}

func (s Screen) focusedOn(target pane) Screen {
	next := s
	next.focus = target

	return next.arranged()
}

func (s Screen) zoomToggled() Screen {
	next := s
	next.zoomed = !s.zoomed

	return next.arranged()
}

func (s Screen) revealing(folderID domain.FolderID) (Screen, []outcome.Outcome) {
	next := s
	next.focus = paneSnippet

	return next.selectingFolder(folderID)
}

func (s Screen) selectingFolder(folderID domain.FolderID) (Screen, []outcome.Outcome) {
	next := s
	next.selectionHolder = paneFolders

	var expanded []outcome.Outcome

	next.panes, expanded = s.panes.selectingFolder(folderID)

	return next.arranged(), expanded
}

func (s Screen) resized(box look.Size) Screen {
	next := s
	next.box = box

	return next.arranged()
}

func (s Screen) arranged() Screen {
	s.layout = s.arrangement()
	s.panes = s.panes.resized(s.layout)

	return s
}

func (s Screen) arrangement() layout {
	if s.zoomed {
		return singlePane(s.box)
	}

	return arrange(s.box, s.focus, s.tallLeft())
}

func (s Screen) tallLeft() pane {
	if s.focus.inLeftColumn() {
		return s.focus
	}

	return s.selectionHolder
}

func (s Screen) hint(hints []key.Binding) string {
	if tooSmall(s.box) && len(hints) == 0 {
		return tooSmallHint
	}

	return hintFor(hints, hintRoom(s.status, s.box.Width))
}

func (s Screen) panesView() string {
	if s.layout.single {
		return s.paneFrame(s.focus)
	}

	left := lipgloss.JoinVertical(lipgloss.Left, s.paneFrame(paneFolders), s.paneFrame(paneTags))

	return lipgloss.JoinHorizontal(lipgloss.Top, left, s.paneFrame(paneList), s.paneFrame(paneSnippet))
}

func (s Screen) paneFrame(p pane) string {
	frame := s.paneStyle(p)

	return look.Frame(frame, p.title(s.panes.selectionAndOrder(s.selection())), s.panes.body(p, frame), s.layout.of(p))
}

func (s Screen) paneStyle(p pane) look.FrameStyle {
	if p == s.focus {
		return s.styles.Focused
	}

	return s.styles.Unfocused
}
