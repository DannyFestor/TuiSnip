package editoverlay_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

func TestSession_UpdateClick(t *testing.T) {
	t.Parallel()

	t.Run("focuses the clicked field", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Click("Description")
		screen.Press(keypress.Typed("typed")...)
		screen.Press(save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: input("", "typed", "")}}, screen.Outcomes())
	})

	t.Run("focuses Tags without opening the Tag editor", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Click("Tags")

		assert.Contains(t, screen.Screen(), "› Tags")
		assert.NotContains(t, screen.Screen(), "filter or new Tag")
	})

	t.Run("focuses Content without entering it on its label", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Click("Content")
		screen.Press(keypress.Typed("lost")...)
		screen.Press(save())

		assert.Contains(t, screen.Screen(), "› Content")
		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: input("", "", "")}}, screen.Outcomes())
	})

	t.Run("enters Content on a click below the fields", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)
		below := screen.CellOf(contentEntryHint)
		below.Y++

		screen.Send(pointer.Clicked{At: below, Double: false})
		screen.Press(keypress.Typed("body")...)
		screen.Press(save())

		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: input("", "", "body")}}, screen.Outcomes())
	})
}

func TestSession_UpdateClickedOutside(t *testing.T) {
	t.Parallel()

	t.Run("closes without changes", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Send(outside())

		assert.False(t, screen.IsOpen())
	})

	t.Run("asks to discard changes", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)
		screen.Press(keypress.Letter('x'))

		screen.Send(outside())

		assert.Contains(t, screen.Screen(), discardQuestion)
	})

	t.Run("leaves Content as esc does", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)
		screen.Press(enterContent()...)

		screen.Send(outside())

		assert.True(t, screen.IsOpen())
		assert.Contains(t, screen.Screen(), contentEntryHint)
	})
}

func outside() pointer.Clicked {
	return pointer.Clicked{At: pointer.Point{X: 0, Y: 0}, Double: false}
}
