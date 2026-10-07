package input_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/input"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
)

func TestLineStyles(t *testing.T) {
	t.Parallel()

	styles := look.NewStyles(look.SchemeLight)
	got := input.LineStyles(styles)

	t.Run("draws the prompt plain", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, styles.Plain, got.Focused.Prompt)
		assert.Equal(t, styles.Plain, got.Blurred.Prompt)
	})

	t.Run("draws the text plain while focused and dim while blurred", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, styles.Plain, got.Focused.Text)
		assert.Equal(t, styles.Dim, got.Blurred.Text)
	})

	t.Run("dims placeholders and suggestions", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, styles.Dim, got.Focused.Placeholder)
		assert.Equal(t, styles.Dim, got.Focused.Suggestion)
		assert.Equal(t, styles.Dim, got.Blurred.Placeholder)
		assert.Equal(t, styles.Dim, got.Blurred.Suggestion)
	})

	t.Run("draws a steady cursor in the accent colour", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, styles.Accent.GetForeground(), got.Cursor.Color)
		assert.False(t, got.Cursor.Blink)
	})
}

func TestAreaStyles(t *testing.T) {
	t.Parallel()

	styles := look.NewStyles(look.SchemeLight)
	got := input.AreaStyles(styles)

	t.Run("draws the text plain while focused and dim while blurred", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, styles.Plain, got.Focused.Text)
		assert.Equal(t, styles.Dim, got.Blurred.Text)
	})

	t.Run("dims line numbers, the placeholder and the end of the buffer", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, styles.Dim, got.Focused.LineNumber)
		assert.Equal(t, styles.Dim, got.Focused.Placeholder)
		assert.Equal(t, styles.Dim, got.Focused.EndOfBuffer)
		assert.Equal(t, styles.Dim, got.Blurred.LineNumber)
		assert.Equal(t, styles.Dim, got.Blurred.Placeholder)
		assert.Equal(t, styles.Dim, got.Blurred.EndOfBuffer)
	})

	t.Run("marks the cursor line only while focused and dims it while blurred", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, styles.CurrentLine, got.Focused.CursorLine)
		assert.Equal(t, styles.Accent, got.Focused.CursorLineNumber)
		assert.Equal(t, styles.Dim, got.Blurred.CursorLine)
		assert.Equal(t, styles.Dim, got.Blurred.CursorLineNumber)
	})

	t.Run("marks the selection whether focused or not", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, styles.Selection, got.Focused.Selection)
		assert.Equal(t, styles.Selection, got.Blurred.Selection)
	})

	t.Run("draws a steady cursor in the accent colour", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, styles.Accent.GetForeground(), got.Cursor.Color)
		assert.False(t, got.Cursor.Blink)
	})
}

func TestRestyledLine(t *testing.T) {
	t.Parallel()

	light := look.NewStyles(look.SchemeLight)
	restyled := input.RestyledLine(input.NewLine("/ ", look.NewStyles(look.SchemeDark)), light)

	assert.Equal(t, input.NewLine("/ ", light).View(), restyled.View())
}

func TestRestyledArea(t *testing.T) {
	t.Parallel()

	light := look.NewStyles(look.SchemeLight)
	restyled := input.RestyledArea(input.NewContentArea(look.NewStyles(look.SchemeDark)), light)

	assert.Equal(t, input.NewContentArea(light).View(), restyled.View())
}
