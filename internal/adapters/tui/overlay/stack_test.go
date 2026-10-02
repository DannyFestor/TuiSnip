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

		stack := stackOf(newFake("bottom"), newFake("top"))

		_, outcomes, _ := stack.Update(letter('a'))

		assert.Equal(t, []overlay.Outcome{heard{by: "top", what: "a"}}, outcomes)
	})

	t.Run("routes a paste to the top overlay only", func(t *testing.T) {
		t.Parallel()

		stack := stackOf(newFake("bottom"), newFake("top"))

		_, outcomes, _ := stack.Update(tea.PasteMsg{Content: "text"})

		assert.Equal(t, []overlay.Outcome{heard{by: "top", what: "paste text"}}, outcomes)
	})

	t.Run("delivers any other message to every overlay, top first", func(t *testing.T) {
		t.Parallel()

		stack := stackOf(newFake("bottom"), newFake("top"))

		_, outcomes, _ := stack.Update(tickMsg{})

		assert.Equal(t, []overlay.Outcome{heard{by: "top", what: "tick"}, heard{by: "bottom", what: "tick"}}, outcomes)
	})

	t.Run("skips an overlay that closed earlier in the same delivery", func(t *testing.T) {
		t.Parallel()

		parent := newFake("parent").receiving(closingOn(heard{by: "child", what: "tick"}))
		stack := stackOf(newFake("bottom"), parent, newFake("child"))

		_, outcomes, _ := stack.Update(tickMsg{})

		assert.Equal(t, []overlay.Outcome{heard{by: "bottom", what: "tick"}}, outcomes)
	})

	t.Run("ignores a key with nothing open", func(t *testing.T) {
		t.Parallel()

		_, outcomes, cmd := overlay.NewStack().Update(letter('a'))

		assert.Empty(t, outcomes)
		assert.Nil(t, cmd)
	})

	t.Run("closes the top overlay that asks to", func(t *testing.T) {
		t.Parallel()

		stack := stackOf(newFake("bottom"), newFake("top").on("c", closing))

		stack, _, _ = stack.Update(letter('c'))

		assert.Equal(t, []string{"bottom"}, hintKeys(stack))
	})

	t.Run("closes when the last overlay closes", func(t *testing.T) {
		t.Parallel()

		stack := stackOf(newFake("only").on("c", closing))

		stack, _, _ = stack.Update(letter('c'))

		assert.False(t, stack.Open())
	})

	t.Run("hands a child's outcome to its parent", func(t *testing.T) {
		t.Parallel()

		stack := stackOf(newFake("parent").receiving(relaying), newFake("child"))

		_, outcomes, _ := stack.Update(letter('a'))

		assert.Equal(t, []overlay.Outcome{relayed{by: "parent", outcome: heard{by: "child", what: "a"}}}, outcomes)
	})

	t.Run("passes an outcome by an overlay that takes none", func(t *testing.T) {
		t.Parallel()

		stack := stackOf(newFake("bottom").receiving(relaying), childlessOverlay{}, newFake("child"))

		_, outcomes, _ := stack.Update(letter('a'))

		assert.Equal(t, []overlay.Outcome{relayed{by: "bottom", outcome: heard{by: "child", what: "a"}}}, outcomes)
	})

	t.Run("keeps an outcome its parent consumes", func(t *testing.T) {
		t.Parallel()

		stack := stackOf(newFake("parent").receiving(staying), newFake("child"))

		_, outcomes, _ := stack.Update(letter('a'))

		assert.Empty(t, outcomes)
	})

	t.Run("closes a child with its parent", func(t *testing.T) {
		t.Parallel()

		parent := newFake("parent").receiving(closingOn(heard{by: "child", what: "a"}))
		stack := stackOf(newFake("bottom"), parent, newFake("child"))
		before := hintKeys(stack)

		stack, _, _ = stack.Update(letter('a'))

		assert.Equal(t, []string{"child"}, before)
		assert.Equal(t, []string{"bottom"}, hintKeys(stack))
	})

	t.Run("opens a child over the overlay that asks to", func(t *testing.T) {
		t.Parallel()

		stack := stackOf(newFake("parent").on("o", opening(newFake("child"))))

		stack, _, _ = stack.Update(letter('o'))

		assert.Equal(t, []string{"child"}, hintKeys(stack))
	})

	t.Run("replaces an overlay's child with the one it opens", func(t *testing.T) {
		t.Parallel()

		parent := newFake("parent").receiving(func(f fakeOverlay, outcome overlay.Outcome) overlay.Step {
			if outcome != (heard{by: "old child", what: "a"}) {
				return overlay.Stay(f)
			}

			return overlay.Stay(f).Opening(newFake("new child").on("c", closing))
		})
		stack := stackOf(parent, newFake("old child"))

		stack, _, _ = stack.Update(letter('a'))
		replaced := hintKeys(stack)
		stack, _, _ = stack.Update(letter('c'))

		assert.Equal(t, []string{"new child"}, replaced)
		assert.Equal(t, []string{"parent"}, hintKeys(stack))
	})

	t.Run("sizes an opened child to the screen", func(t *testing.T) {
		t.Parallel()

		stack := stackOf(newFake("parent").on("o", opening(newFake("child"))))
		stack, _, _ = stack.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

		_, outcomes, _ := stack.Update(letter('o'))

		assert.Equal(t, []overlay.Outcome{heard{by: "child", what: "80x24"}}, outcomes)
	})

	t.Run("runs the commands overlays return", func(t *testing.T) {
		t.Parallel()

		ticking := func(f fakeOverlay) overlay.Step {
			return overlay.Stay(f).Running(func() tea.Msg { return tickMsg{} })
		}
		stack := stackOf(newFake("only").on("r", ticking))

		_, _, cmd := stack.Update(letter('r'))

		require.NotNil(t, cmd)
		assert.Equal(t, tickMsg{}, cmd())
	})
}

func TestStack_Offered(t *testing.T) {
	t.Parallel()

	t.Run("offers an outcome to the top overlay, then down the stack", func(t *testing.T) {
		t.Parallel()

		stack := stackOf(newFake("bottom").receiving(relaying), newFake("top").receiving(relaying))

		_, outcomes, _ := stack.Offered("quit")

		want := relayed{by: "bottom", outcome: relayed{by: "top", outcome: "quit"}}
		assert.Equal(t, []overlay.Outcome{want}, outcomes)
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

		stack, _, _ := stackOf(newFake("bottom").looking("BBBBBB"), newFake("top").looking("TT")).
			Update(tea.WindowSizeMsg{Width: 10, Height: 1})

		assert.Equal(t, "..BBTTBB..", stack.Render(".........."))
	})

	t.Run("keeps the background with nothing open", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "background", overlay.NewStack().Render("background"))
	})
}

func TestStack_Pushed(t *testing.T) {
	t.Parallel()

	t.Run("sizes the pushed overlay to the last screen size", func(t *testing.T) {
		t.Parallel()

		stack, _, _ := overlay.NewStack().Update(tea.WindowSizeMsg{Width: 80, Height: 24})

		_, outcomes, _ := stack.Pushed(newFake("pushed"))

		assert.Equal(t, []overlay.Outcome{heard{by: "pushed", what: "80x24"}}, outcomes)
	})

	t.Run("opens the stack", func(t *testing.T) {
		t.Parallel()

		assert.False(t, overlay.NewStack().Open())
		assert.True(t, stackOf(newFake("pushed")).Open())
	})
}

func TestStack_Hints(t *testing.T) {
	t.Parallel()

	t.Run("shows the top overlay's hints", func(t *testing.T) {
		t.Parallel()

		stack := stackOf(newFake("bottom"), newFake("top"))

		assert.Equal(t, []string{"top"}, hintKeys(stack))
	})

	t.Run("shows none with nothing open", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, overlay.NewStack().Hints())
	})
}
