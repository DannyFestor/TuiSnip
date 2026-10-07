package folderpicker_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
)

func TestPicker_UpdateClick(t *testing.T) {
	t.Parallel()

	t.Run("picks the clicked Folder and closes", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		screen := picking(t, sample, domain.FolderID{}, domain.FolderID{})

		screen.Click("go / testing")

		assert.Equal(t, []outcome.Outcome{pickedOutcome(sample.Tests.ID())}, screen.Outcomes())
		assert.False(t, screen.IsOpen())
	})

	t.Run("closes without picking on a click outside", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, foldertree.New(t), domain.FolderID{}, domain.FolderID{})

		screen.Send(pointer.Clicked{At: pointer.Point{X: 0, Y: 0}, Double: false})

		assert.Empty(t, screen.Outcomes())
		assert.False(t, screen.IsOpen())
	})
}
