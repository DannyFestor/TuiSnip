package picker_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/picker"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/test/screencell"
)

func TestList_UpdateClick(t *testing.T) {
	t.Parallel()

	t.Run("picks the clicked choice", func(t *testing.T) {
		t.Parallel()

		list := opened(t, choicesOf("Go", "Bash", "YAML"))

		_, result, _ := list.Update(clickOn(t, list, "Bash"))

		assert.Equal(t, picked(1), result)
	})

	t.Run("picks a clicked choice among those the filter shows", func(t *testing.T) {
		t.Parallel()

		list := typed(t, opened(t, choicesOf("Go", "Bash", "YAML")), "a")

		_, result, _ := list.Update(clickOn(t, list, "YAML"))

		assert.Equal(t, picked(2), result)
	})

	t.Run("picks nothing on a greyed-out choice", func(t *testing.T) {
		t.Parallel()

		list := opened(t, choicesOf("Go", "Bash")).WithGreyedOut(1)

		_, result, _ := list.Update(clickOn(t, list, "Bash"))

		assert.Equal(t, picker.Filtering, result.Ending)
	})

	t.Run("picks nothing on the filter line", func(t *testing.T) {
		t.Parallel()

		list := opened(t, choicesOf("Go", "Bash"))

		_, result, _ := list.Update(clickOn(t, list, "filter:"))

		assert.Equal(t, picker.Filtering, result.Ending)
	})

	t.Run("picks nothing below the last choice", func(t *testing.T) {
		t.Parallel()

		list := opened(t, choicesOf("Go", "Bash"))
		below := clickOn(t, list, "Bash")
		below.At.Y++

		_, result, _ := list.Update(below)

		assert.Equal(t, picker.Filtering, result.Ending)
	})
}

func TestList_UpdateClickedOutside(t *testing.T) {
	t.Parallel()

	_, result, _ := opened(t, choicesOf("Go")).Update(pointer.ClickedOutside{})

	assert.Equal(t, picker.Cancelled, result.Ending)
}

func clickOn(t *testing.T, list picker.List, text string) pointer.Clicked {
	t.Helper()

	return pointer.Clicked{At: screencell.Find(t, list.View(), text), Double: false}
}
