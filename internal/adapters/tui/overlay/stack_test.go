package overlay_test

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

func TestStack_Update(t *testing.T) {
	t.Parallel()

	t.Run("routes a key to the top overlay only", func(t *testing.T) {
		t.Parallel()

		top := overlayMock(t)
		top.EXPECT().Update(keypress.Letter('a')).Return(stay(top).Passing("top pressed"))
		stack := stackOf(overlayMock(t), top)

		_, outcomes, _ := stack.Update(keypress.Letter('a'))

		assert.Equal(t, []string{"top pressed"}, outcomes)
	})

	t.Run("routes a paste to the top overlay only", func(t *testing.T) {
		t.Parallel()

		paste := tea.PasteMsg{Content: "text"}
		top := overlayMock(t)
		top.EXPECT().Update(paste).Return(stay(top).Passing("top pasted"))
		stack := stackOf(overlayMock(t), top)

		_, outcomes, _ := stack.Update(paste)

		assert.Equal(t, []string{"top pasted"}, outcomes)
	})

	t.Run("routes a click to the top overlay only, from its own corner", func(t *testing.T) {
		t.Parallel()

		top := overlayMock(t)
		top.EXPECT().View().Return("AB\nCD")
		top.EXPECT().Update(clickAt(1, 1)).Return(stay(top).Passing("top clicked"))
		stack := smallScreen(t, overlayMock(t), top)

		_, outcomes, _ := stack.Update(clickAt(5, 1))

		assert.Equal(t, []string{"top clicked"}, outcomes)
	})

	t.Run("tells the top overlay of a click outside it", func(t *testing.T) {
		t.Parallel()

		top := overlayMock(t)
		top.EXPECT().View().Return("AB\nCD")
		top.EXPECT().Update(pointer.ClickedOutside{}).Return(stay(top).Passing("top dismissed"))
		stack := smallScreen(t, overlayMock(t), top)

		_, outcomes, _ := stack.Update(clickAt(3, 1))

		assert.Equal(t, []string{"top dismissed"}, outcomes)
	})

	t.Run("routes a click on the base unchanged", func(t *testing.T) {
		t.Parallel()

		base := baseMock(t)
		base.EXPECT().Update(clickAt(3, 1)).Return(stay(base).Passing("base clicked"))
		stack := smallScreen(t, base)

		_, outcomes, _ := stack.Update(clickAt(3, 1))

		assert.Equal(t, []string{"base clicked"}, outcomes)
	})

	t.Run("routes a wheel turn over the top overlay, from its own corner", func(t *testing.T) {
		t.Parallel()

		wheeled := pointer.Wheeled{At: pointer.Point{X: 5, Y: 1}, Lines: 3}
		top := overlayMock(t)
		top.EXPECT().View().Return("AB\nCD")
		top.EXPECT().Update(wheeled.Relative(pointer.Point{X: 4, Y: 0})).Return(stay(top).Passing("top wheeled"))
		stack := smallScreen(t, overlayMock(t), top)

		_, outcomes, _ := stack.Update(wheeled)

		assert.Equal(t, []string{"top wheeled"}, outcomes)
	})

	t.Run("drops a wheel turn outside the top overlay", func(t *testing.T) {
		t.Parallel()

		top := overlayMock(t)
		top.EXPECT().View().Return("AB\nCD")
		stack := smallScreen(t, overlayMock(t), top)

		_, outcomes, _ := stack.Update(pointer.Wheeled{At: pointer.Point{X: 0, Y: 0}, Lines: 3})

		assert.Empty(t, outcomes)
	})

	t.Run("ignores a click with nothing open", func(t *testing.T) {
		t.Parallel()

		_, outcomes, _ := emptyStack().Update(clickAt(0, 0))

		assert.Empty(t, outcomes)
	})

	t.Run("delivers any other message to every overlay, top first", func(t *testing.T) {
		t.Parallel()

		bottom := overlayMock(t)
		bottom.EXPECT().Update(tickMsg{}).Return(stay(bottom).Passing("bottom ticked"))

		top := overlayMock(t)
		top.EXPECT().Update(tickMsg{}).Return(stay(top).Passing("top ticked"))
		stack := stackOf(bottom, top)

		_, outcomes, _ := stack.Update(tickMsg{})

		assert.Equal(t, []string{"top ticked", "bottom ticked"}, outcomes)
	})

	t.Run("skips an overlay that closed earlier in the same delivery", func(t *testing.T) {
		t.Parallel()

		bottom := overlayMock(t)
		bottom.EXPECT().Update(tickMsg{}).Return(stay(bottom).Passing("bottom ticked"))

		parent := parentMock(t)
		parent.EXPECT().Received("child ticked").Return(closing())

		child := overlayMock(t)
		child.EXPECT().Update(tickMsg{}).Return(stay(child).Passing("child ticked"))
		stack := stackOf(bottom, parent, child)

		_, outcomes, _ := stack.Update(tickMsg{})

		assert.Equal(t, []string{"bottom ticked"}, outcomes)
	})

	t.Run("ignores a key with nothing open", func(t *testing.T) {
		t.Parallel()

		_, outcomes, cmd := emptyStack().Update(keypress.Letter('a'))

		assert.Empty(t, outcomes)
		assert.Nil(t, cmd)
	})

	t.Run("closes the top overlay that asks to", func(t *testing.T) {
		t.Parallel()

		bottom := overlayMock(t)
		bottom.EXPECT().ShortHelp().Return(hinting("bottom"))

		top := overlayMock(t)
		top.EXPECT().Update(keypress.Letter('c')).Return(closing())
		stack := stackOf(bottom, top)

		stack, _, _ = stack.Update(keypress.Letter('c'))

		assert.Equal(t, []string{"bottom"}, hintKeys(stack))
	})

	t.Run("closes when the last overlay closes", func(t *testing.T) {
		t.Parallel()

		only := overlayMock(t)
		only.EXPECT().Update(keypress.Letter('c')).Return(closing())
		stack := stackOf(only)

		stack, _, _ = stack.Update(keypress.Letter('c'))

		assert.False(t, stack.Open())
	})

	t.Run("hands a child's outcome to its parent", func(t *testing.T) {
		t.Parallel()

		parent := parentMock(t)
		parent.EXPECT().Received("picked").Return(stay(parent).Passing("saved"))

		child := overlayMock(t)
		child.EXPECT().Update(keypress.Letter('a')).Return(stay(child).Passing("picked"))
		stack := stackOf(parent, child)

		_, outcomes, _ := stack.Update(keypress.Letter('a'))

		assert.Equal(t, []string{"saved"}, outcomes)
	})

	t.Run("hands on a zero-value outcome", func(t *testing.T) {
		t.Parallel()

		only := overlayMock(t)
		only.EXPECT().Update(keypress.Letter('a')).Return(stay(only).Passing(""))
		stack := stackOf(only)

		_, outcomes, _ := stack.Update(keypress.Letter('a'))

		assert.Equal(t, []string{""}, outcomes)
	})

	t.Run("hands on every outcome a step passes, in order", func(t *testing.T) {
		t.Parallel()

		only := overlayMock(t)
		only.EXPECT().Update(keypress.Letter('a')).Return(stay(only).Passing("first").Passing("second"))
		stack := stackOf(only)

		_, outcomes, _ := stack.Update(keypress.Letter('a'))

		assert.Equal(t, []string{"first", "second"}, outcomes)
	})

	t.Run("hands the later outcomes past a parent that closed on an earlier one", func(t *testing.T) {
		t.Parallel()

		parent := parentMock(t)
		parent.EXPECT().Received("first").Return(closing())

		child := overlayMock(t)
		child.EXPECT().Update(keypress.Letter('a')).Return(stay(child).Passing("first").Passing("second"))
		stack := stackOf(parent, child)

		_, outcomes, _ := stack.Update(keypress.Letter('a'))

		assert.Equal(t, []string{"second"}, outcomes)
	})

	t.Run("passes an outcome by an overlay that takes none", func(t *testing.T) {
		t.Parallel()

		parent := parentMock(t)
		parent.EXPECT().Received("picked").Return(stay(parent).Passing("saved"))

		child := overlayMock(t)
		child.EXPECT().Update(keypress.Letter('a')).Return(stay(child).Passing("picked"))
		stack := stackOf(parent, overlayMock(t), child)

		_, outcomes, _ := stack.Update(keypress.Letter('a'))

		assert.Equal(t, []string{"saved"}, outcomes)
	})

	t.Run("keeps an outcome its parent consumes", func(t *testing.T) {
		t.Parallel()

		parent := parentMock(t)
		parent.EXPECT().Received("picked").Return(stay(parent))

		child := overlayMock(t)
		child.EXPECT().Update(keypress.Letter('a')).Return(stay(child).Passing("picked"))
		stack := stackOf(parent, child)

		_, outcomes, _ := stack.Update(keypress.Letter('a'))

		assert.Empty(t, outcomes)
	})

	t.Run("closes a child with its parent", func(t *testing.T) {
		t.Parallel()

		bottom := overlayMock(t)
		bottom.EXPECT().ShortHelp().Return(hinting("bottom"))

		parent := parentMock(t)
		parent.EXPECT().Received("done").Return(closing())

		child := overlayMock(t)
		child.EXPECT().Update(keypress.Letter('a')).Return(stay(child).Passing("done"))
		stack := stackOf(bottom, parent, child)

		stack, _, _ = stack.Update(keypress.Letter('a'))

		assert.Equal(t, []string{"bottom"}, hintKeys(stack))
	})

	t.Run("opens a child over the overlay that asks to", func(t *testing.T) {
		t.Parallel()

		child := overlayMock(t)
		child.EXPECT().ShortHelp().Return(hinting("child"))

		parent := overlayMock(t)
		parent.EXPECT().Update(keypress.Letter('o')).Return(stay(parent).Opening(child))
		stack := stackOf(parent)

		stack, _, _ = stack.Update(keypress.Letter('o'))

		assert.Equal(t, []string{"child"}, hintKeys(stack))
	})

	t.Run("replaces an overlay's child with the one it opens", func(t *testing.T) {
		t.Parallel()

		newChild := overlayMock(t)
		newChild.EXPECT().ShortHelp().Return(hinting("new child"))
		newChild.EXPECT().Update(keypress.Letter('c')).Return(closing())

		parent := parentMock(t)
		parent.EXPECT().Received("asked").Return(stay(parent).Opening(newChild))
		parent.EXPECT().ShortHelp().Return(hinting("parent"))

		oldChild := overlayMock(t)
		oldChild.EXPECT().Update(keypress.Letter('a')).Return(stay(oldChild).Passing("asked"))
		stack := stackOf(parent, oldChild)

		stack, _, _ = stack.Update(keypress.Letter('a'))
		replaced := hintKeys(stack)
		stack, _, _ = stack.Update(keypress.Letter('c'))

		assert.Equal(t, []string{"new child"}, replaced)
		assert.Equal(t, []string{"parent"}, hintKeys(stack))
	})

	t.Run("resizes every overlay to the whole screen", func(t *testing.T) {
		t.Parallel()

		stackOf(terminalSizedMock(t), terminalSizedMock(t)).Update(terminal())
	})

	t.Run("resizes an opened child to the whole screen", func(t *testing.T) {
		t.Parallel()

		child := NewMockOverlay[string](t)
		child.EXPECT().Update(wholeTerminal()).Return(stay(child)).Once()

		parent := overlayMock(t)
		parent.EXPECT().Update(keypress.Letter('o')).Return(stay(parent).Opening(child))
		stack, _, _ := stackOf(parent).Update(terminal())

		stack.Update(keypress.Letter('o'))
	})

	t.Run("runs the commands overlays return", func(t *testing.T) {
		t.Parallel()

		only := overlayMock(t)
		only.EXPECT().Update(keypress.Letter('r')).Return(stay(only).Running(func() tea.Msg { return tickMsg{} }))
		stack := stackOf(only)

		_, _, cmd := stack.Update(keypress.Letter('r'))

		require.NotNil(t, cmd)
		assert.Equal(t, tickMsg{}, cmd())
	})
}

func TestStack_Pushed(t *testing.T) {
	t.Parallel()

	t.Run("resizes the pushed overlay to the last screen size", func(t *testing.T) {
		t.Parallel()

		pushed := NewMockOverlay[string](t)
		pushed.EXPECT().Update(wholeTerminal()).Return(stay(pushed)).Once()
		stack, _, _ := emptyStack().Update(terminal())

		stack.Pushed(pushed)
	})

	t.Run("opens the stack", func(t *testing.T) {
		t.Parallel()

		assert.False(t, emptyStack().Open())
		assert.True(t, stackOf(overlayMock(t)).Open())
	})
}

func TestStack_ShortHelp(t *testing.T) {
	t.Parallel()

	t.Run("shows the top overlay's short help", func(t *testing.T) {
		t.Parallel()

		top := overlayMock(t)
		top.EXPECT().ShortHelp().Return(hinting("top"))
		stack := stackOf(overlayMock(t), top)

		assert.Equal(t, []string{"top"}, hintKeys(stack))
	})

	t.Run("shows none with nothing open", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, emptyStack().ShortHelp())
	})
}

func TestStack_FullHelp(t *testing.T) {
	t.Parallel()

	t.Run("shows the top overlay's full help", func(t *testing.T) {
		t.Parallel()

		full := [][]key.Binding{hinting("top"), hinting("more")}
		top := overlayMock(t)
		top.EXPECT().FullHelp().Return(full)
		stack := stackOf(overlayMock(t), top)

		assert.Equal(t, full, stack.FullHelp())
	})

	t.Run("shows none with nothing open", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, emptyStack().FullHelp())
	})
}

func TestStack_Offered(t *testing.T) {
	t.Parallel()

	t.Run("offers an outcome to the top overlay, then down the stack", func(t *testing.T) {
		t.Parallel()

		bottom := parentMock(t)
		bottom.EXPECT().Received("top passed quit").Return(stay(bottom).Passing("bottom passed it on"))

		top := parentMock(t)
		top.EXPECT().Received("quit").Return(stay(top).Passing("top passed quit"))
		stack := stackOf(bottom, top)

		_, outcomes, _ := stack.Offered("quit")

		assert.Equal(t, []string{"bottom passed it on"}, outcomes)
	})

	t.Run("returns the outcome with nothing open", func(t *testing.T) {
		t.Parallel()

		_, outcomes, _ := emptyStack().Offered("quit")

		assert.Equal(t, []string{"quit"}, outcomes)
	})
}

func TestStack_Render(t *testing.T) {
	t.Parallel()

	t.Run("draws a base under the overlays, showing the top one's hints", func(t *testing.T) {
		t.Parallel()

		top := overlayMock(t)
		top.EXPECT().View().Return("TT")
		top.EXPECT().ShortHelp().Return(hinting("top"))

		base := baseMock(t)
		base.EXPECT().ViewUnder(hinting("top")).Return("..........")
		stack, _, _ := stackOf(base, top).Update(tea.WindowSizeMsg{Width: 10, Height: 1})

		assert.Equal(t, "....TT....", stack.Render())
	})

	t.Run("draws the overlays centred over each other, bottom first", func(t *testing.T) {
		t.Parallel()

		middle := overlayMock(t)
		middle.EXPECT().View().Return("MMMMMM")

		top := overlayMock(t)
		top.EXPECT().View().Return("TT")
		top.EXPECT().ShortHelp().Return(nil)

		base := baseMock(t)
		base.EXPECT().ViewUnder(mock.Anything).Return("..........")
		stack, _, _ := stackOf(base, middle, top).Update(tea.WindowSizeMsg{Width: 10, Height: 1})

		assert.Equal(t, "..MMTTMM..", stack.Render())
	})

	t.Run("centres an overlay vertically", func(t *testing.T) {
		t.Parallel()

		top := overlayMock(t)
		top.EXPECT().View().Return("T")
		top.EXPECT().ShortHelp().Return(nil)

		base := baseMock(t)
		base.EXPECT().ViewUnder(mock.Anything).Return("...\n...\n...")
		stack, _, _ := stackOf(base, top).Update(tea.WindowSizeMsg{Width: 3, Height: 3})

		assert.Equal(t, "...\n.T.\n...", stack.Render())
	})

	t.Run("draws nothing with nothing open", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, emptyStack().Render())
	})
}
