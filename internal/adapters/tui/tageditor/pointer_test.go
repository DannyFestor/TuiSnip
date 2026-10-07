package tageditor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagchoice"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

func TestEditor_UpdateClick(t *testing.T) {
	t.Parallel()

	t.Run("toggles the clicked Tag and stays open", func(t *testing.T) {
		t.Parallel()

		tags := newSampleTags(t)
		screen := editing(t, tags.listed(), tagchoice.Chosen{})

		screen.Click("testing")

		assert.True(t, screen.IsOpen())
		assert.Equal(t, []domain.Tag{tags.testing}, edited(t, screen).Stored())
	})

	t.Run("closes with the chosen Tags on a click outside", func(t *testing.T) {
		t.Parallel()

		tags := newSampleTags(t)
		screen := editing(t, tags.listed(), tagchoice.Of([]domain.Tag{tags.docker}))

		screen.Send(pointer.Clicked{At: pointer.Point{X: 0, Y: 0}, Double: false})

		assert.False(t, screen.IsOpen())
		require.Len(t, screen.Outcomes(), 1)
		reported, ok := screen.Outcomes()[0].(outcome.TagsEdited)
		require.True(t, ok)
		assert.Equal(t, []domain.Tag{tags.docker}, reported.Chosen.Stored())
	})
}
