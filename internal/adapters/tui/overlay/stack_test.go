package overlay_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/overlay"
)

func TestStack_Update(t *testing.T) {
	t.Parallel()

	t.Run("routes a key to the top overlay only", func(t *testing.T) {
		t.Parallel()

		top := overlayMock(t)
		top.EXPECT().Update(letter('a')).Return(overlay.Stay(top).Passing("top pressed"))
		stack := stackOf(overlayMock(t), top)

		_, outcomes, _ := stack.Update(letter('a'))

		assert.Equal(t, []overlay.Outcome{"top pressed"}, outcomes)
	})

	t.Run("routes a paste to the top overlay only", func(t *testing.T) {
		t.Parallel()

		paste := tea.PasteMsg{Content: "text"}
		top := overlayMock(t)
		top.EXPECT().Update(paste).Return(overlay.Stay(top).Passing("top pasted"))
		stack := stackOf(overlayMock(t), top)

		_, outcomes, _ := stack.Update(paste)

		assert.Equal(t, []overlay.Outcome{"top pasted"}, outcomes)
	})

	t.Run("delivers any other message to every overlay, top first", func(t *testing.T) {
		t.Parallel()

		bottom := overlayMock(t)
		bottom.EXPECT().Update(tickMsg{}).Return(overlay.Stay(bottom).Passing("bottom ticked"))

		top := overlayMock(t)
		top.EXPECT().Update(tickMsg{}).Return(overlay.Stay(top).Passing("top ticked"))
		stack := stackOf(bottom, top)

		_, outcomes, _ := stack.Update(tickMsg{})

		assert.Equal(t, []overlay.Outcome{"top ticked", "bottom ticked"}, outcomes)
	})

	t.Run("skips an overlay that closed earlier in the same delivery", func(t *testing.T) {
		t.Parallel()

		bottom := overlayMock(t)
		bottom.EXPECT().Update(tickMsg{}).Return(overlay.Stay(bottom).Passing("bottom ticked"))

		parent := parentMock(t)
		parent.EXPECT().Received("child ticked").Return(overlay.Close())

		child := overlayMock(t)
		child.EXPECT().Update(tickMsg{}).Return(overlay.Stay(child).Passing("child ticked"))
		stack := stackOf(bottom, parent, child)

		_, outcomes, _ := stack.Update(tickMsg{})

		assert.Equal(t, []overlay.Outcome{"bottom ticked"}, outcomes)
	})

	t.Run("ignores a key with nothing open", func(t *testing.T) {
		t.Parallel()

		_, outcomes, cmd := overlay.NewStack().Update(letter('a'))

		assert.Empty(t, outcomes)
		assert.Nil(t, cmd)
	})

	t.Run("closes the top overlay that asks to", func(t *testing.T) {
		t.Parallel()

		bottom := overlayMock(t)
		bottom.EXPECT().Hints().Return(hinting("bottom"))

		top := overlayMock(t)
		top.EXPECT().Update(letter('c')).Return(overlay.Close())
		stack := stackOf(bottom, top)

		stack, _, _ = stack.Update(letter('c'))

		assert.Equal(t, []string{"bottom"}, hintKeys(stack))
	})

	t.Run("closes when the last overlay closes", func(t *testing.T) {
		t.Parallel()

		only := overlayMock(t)
		only.EXPECT().Update(letter('c')).Return(overlay.Close())
		stack := stackOf(only)

		stack, _, _ = stack.Update(letter('c'))

		assert.False(t, stack.Open())
	})

	t.Run("hands a child's outcome to its parent", func(t *testing.T) {
		t.Parallel()

		parent := parentMock(t)
		parent.EXPECT().Received("picked").Return(overlay.Stay(parent).Passing("saved"))

		child := overlayMock(t)
		child.EXPECT().Update(letter('a')).Return(overlay.Stay(child).Passing("picked"))
		stack := stackOf(parent, child)

		_, outcomes, _ := stack.Update(letter('a'))

		assert.Equal(t, []overlay.Outcome{"saved"}, outcomes)
	})

	t.Run("passes an outcome by an overlay that takes none", func(t *testing.T) {
		t.Parallel()

		parent := parentMock(t)
		parent.EXPECT().Received("picked").Return(overlay.Stay(parent).Passing("saved"))

		child := overlayMock(t)
		child.EXPECT().Update(letter('a')).Return(overlay.Stay(child).Passing("picked"))
		stack := stackOf(parent, overlayMock(t), child)

		_, outcomes, _ := stack.Update(letter('a'))

		assert.Equal(t, []overlay.Outcome{"saved"}, outcomes)
	})

	t.Run("keeps an outcome its parent consumes", func(t *testing.T) {
		t.Parallel()

		parent := parentMock(t)
		parent.EXPECT().Received("picked").Return(overlay.Stay(parent))

		child := overlayMock(t)
		child.EXPECT().Update(letter('a')).Return(overlay.Stay(child).Passing("picked"))
		stack := stackOf(parent, child)

		_, outcomes, _ := stack.Update(letter('a'))

		assert.Empty(t, outcomes)
	})

	t.Run("closes a child with its parent", func(t *testing.T) {
		t.Parallel()

		bottom := overlayMock(t)
		bottom.EXPECT().Hints().Return(hinting("bottom"))

		parent := parentMock(t)
		parent.EXPECT().Received("done").Return(overlay.Close())

		child := overlayMock(t)
		child.EXPECT().Update(letter('a')).Return(overlay.Stay(child).Passing("done"))
		stack := stackOf(bottom, parent, child)

		stack, _, _ = stack.Update(letter('a'))

		assert.Equal(t, []string{"bottom"}, hintKeys(stack))
	})

	t.Run("opens a child over the overlay that asks to", func(t *testing.T) {
		t.Parallel()

		child := overlayMock(t)
		child.EXPECT().Hints().Return(hinting("child"))

		parent := overlayMock(t)
		parent.EXPECT().Update(letter('o')).Return(overlay.Stay(parent).Opening(child))
		stack := stackOf(parent)

		stack, _, _ = stack.Update(letter('o'))

		assert.Equal(t, []string{"child"}, hintKeys(stack))
	})

	t.Run("replaces an overlay's child with the one it opens", func(t *testing.T) {
		t.Parallel()

		newChild := overlayMock(t)
		newChild.EXPECT().Hints().Return(hinting("new child"))
		newChild.EXPECT().Update(letter('c')).Return(overlay.Close())

		parent := parentMock(t)
		parent.EXPECT().Received("asked").Return(overlay.Stay(parent).Opening(newChild))
		parent.EXPECT().Hints().Return(hinting("parent"))

		oldChild := overlayMock(t)
		oldChild.EXPECT().Update(letter('a')).Return(overlay.Stay(oldChild).Passing("asked"))
		stack := stackOf(parent, oldChild)

		stack, _, _ = stack.Update(letter('a'))
		replaced := hintKeys(stack)
		stack, _, _ = stack.Update(letter('c'))

		assert.Equal(t, []string{"new child"}, replaced)
		assert.Equal(t, []string{"parent"}, hintKeys(stack))
	})

	t.Run("sizes an opened child to the screen", func(t *testing.T) {
		t.Parallel()

		screen := tea.WindowSizeMsg{Width: 80, Height: 24}
		child := NewMockOverlay(t)
		child.EXPECT().Update(screen).Return(overlay.Stay(child)).Once()

		parent := overlayMock(t)
		parent.EXPECT().Update(letter('o')).Return(overlay.Stay(parent).Opening(child))
		stack, _, _ := stackOf(parent).Update(screen)

		stack.Update(letter('o'))
	})

	t.Run("runs the commands overlays return", func(t *testing.T) {
		t.Parallel()

		only := overlayMock(t)
		only.EXPECT().Update(letter('r')).Return(overlay.Stay(only).Running(func() tea.Msg { return tickMsg{} }))
		stack := stackOf(only)

		_, _, cmd := stack.Update(letter('r'))

		require.NotNil(t, cmd)
		assert.Equal(t, tickMsg{}, cmd())
	})
}

func TestStack_Pushed(t *testing.T) {
	t.Parallel()

	t.Run("sizes the pushed overlay to the last screen size", func(t *testing.T) {
		t.Parallel()

		screen := tea.WindowSizeMsg{Width: 80, Height: 24}
		pushed := NewMockOverlay(t)
		pushed.EXPECT().Update(screen).Return(overlay.Stay(pushed)).Once()
		stack, _, _ := overlay.NewStack().Update(screen)

		stack.Pushed(pushed)
	})

	t.Run("opens the stack", func(t *testing.T) {
		t.Parallel()

		assert.False(t, overlay.NewStack().Open())
		assert.True(t, stackOf(overlayMock(t)).Open())
	})
}

func TestStack_Hints(t *testing.T) {
	t.Parallel()

	t.Run("shows the top overlay's hints", func(t *testing.T) {
		t.Parallel()

		top := overlayMock(t)
		top.EXPECT().Hints().Return(hinting("top"))
		stack := stackOf(overlayMock(t), top)

		assert.Equal(t, []string{"top"}, hintKeys(stack))
	})

	t.Run("shows none with nothing open", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, overlay.NewStack().Hints())
	})
}

func TestStack_Offered(t *testing.T) {
	t.Parallel()

	t.Run("offers an outcome to the top overlay, then down the stack", func(t *testing.T) {
		t.Parallel()

		bottom := parentMock(t)
		bottom.EXPECT().Received("top passed quit").Return(overlay.Stay(bottom).Passing("bottom passed it on"))

		top := parentMock(t)
		top.EXPECT().Received("quit").Return(overlay.Stay(top).Passing("top passed quit"))
		stack := stackOf(bottom, top)

		_, outcomes, _ := stack.Offered("quit")

		assert.Equal(t, []overlay.Outcome{"bottom passed it on"}, outcomes)
	})

	t.Run("returns the outcome with nothing open", func(t *testing.T) {
		t.Parallel()

		_, outcomes, _ := overlay.NewStack().Offered("quit")

		assert.Equal(t, []overlay.Outcome{"quit"}, outcomes)
	})
}

func TestStack_Render(t *testing.T) {
	t.Parallel()

	t.Run("draws the overlays centred over the background, bottom first", func(t *testing.T) {
		t.Parallel()

		bottom := overlayMock(t)
		bottom.EXPECT().View().Return("BBBBBB")

		top := overlayMock(t)
		top.EXPECT().View().Return("TT")
		stack, _, _ := stackOf(bottom, top).Update(tea.WindowSizeMsg{Width: 10, Height: 1})

		assert.Equal(t, "..BBTTBB..", stack.Render(".........."))
	})

	t.Run("centres an overlay vertically", func(t *testing.T) {
		t.Parallel()

		only := overlayMock(t)
		only.EXPECT().View().Return("T")
		stack, _, _ := stackOf(only).Update(tea.WindowSizeMsg{Width: 3, Height: 3})

		assert.Equal(t, "...\n.T.\n...", stack.Render("...\n...\n..."))
	})

	t.Run("keeps the background with nothing open", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "background", overlay.NewStack().Render("background"))
	})
}
