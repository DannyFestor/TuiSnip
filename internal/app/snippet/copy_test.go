package snippet_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

type copyConstructor func(snippet.Finder, snippet.Copier) (*snippet.Copy, error)

func TestNewCopy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		constructor copyConstructor
		wantPrefix  string
	}{
		{name: "keeping content", constructor: snippet.NewCopy, wantPrefix: "snippet.NewCopy: "},
		{
			name:        "trimming the trailing newline",
			constructor: snippet.NewCopyTrimmingTrailingNewline,
			wantPrefix:  "snippet.NewCopyTrimmingTrailingNewline: ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name+" names every missing dependency", func(t *testing.T) {
			t.Parallel()

			_, err := tt.constructor(nil, nil)

			require.ErrorIs(t, err, domain.ErrMissingDependency)
			require.ErrorContains(t, err, tt.wantPrefix)
			require.ErrorContains(t, err, "finder")
			require.ErrorContains(t, err, "copier")
		})
	}
}

func TestCopy_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		constructor copyConstructor
		content     string
		wantText    string
	}{
		{
			name:        "copies the content as stored",
			constructor: snippet.NewCopy,
			content:     "echo hi\n",
			wantText:    "echo hi\n",
		},
		{
			name:        "copies the content without its trailing newline",
			constructor: snippet.NewCopyTrimmingTrailingNewline,
			content:     "echo hi\n",
			wantText:    "echo hi",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stored := testkit.Snippet(t, testkit.SnippetSpec{Fragment: testkit.FragmentSpec{Content: tt.content}})
			finder := NewMockFinder(t)
			finder.EXPECT().Find(mock.Anything, stored.ID()).Return(stored, nil)

			copier := NewMockCopier(t)
			copier.EXPECT().Copy(mock.Anything, tt.wantText).Return(domain.CopyDeliverySentToTerminal, nil)

			result, err := newCopy(t, tt.constructor, finder, copier).Run(
				t.Context(), snippet.CopyInput{SnippetID: stored.ID()},
			)

			require.NoError(t, err)
			assert.Equal(t, domain.CopyDeliverySentToTerminal, result.Delivery)
			assert.Equal(t, tt.wantText, result.Content.String())
		})
	}
}

func TestCopy_RunFailures(t *testing.T) {
	t.Parallel()

	stored := testkit.Snippet(t, testkit.SnippetSpec{})

	t.Run("returns a missing Snippet without copying", func(t *testing.T) {
		t.Parallel()

		finder := NewMockFinder(t)
		finder.EXPECT().Find(mock.Anything, stored.ID()).Return(domain.Snippet{}, domain.ErrNotFound)

		_, err := newCopy(t, snippet.NewCopy, finder, NewMockCopier(t)).Run(
			t.Context(), snippet.CopyInput{SnippetID: stored.ID()},
		)

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "snippet.Copy: ")
	})

	t.Run("returns the clipboard error", func(t *testing.T) {
		t.Parallel()

		finder := NewMockFinder(t)
		finder.EXPECT().Find(mock.Anything, stored.ID()).Return(stored, nil)

		copier := NewMockCopier(t)
		copier.EXPECT().Copy(mock.Anything, mock.Anything).Return("", domain.ErrNoClipboardTool)

		_, err := newCopy(t, snippet.NewCopy, finder, copier).Run(
			t.Context(), snippet.CopyInput{SnippetID: stored.ID()},
		)

		require.ErrorIs(t, err, domain.ErrNoClipboardTool)
		assert.ErrorContains(t, err, "snippet.Copy: ")
	})
}

func newCopy(
	t *testing.T, constructor copyConstructor, finder snippet.Finder, copier snippet.Copier,
) *snippet.Copy {
	t.Helper()

	copyAction, err := constructor(finder, copier)
	require.NoError(t, err)

	return copyAction
}
