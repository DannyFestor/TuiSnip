package testkit_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestSnippet_Defaults(t *testing.T) {
	t.Parallel()

	snippet := testkit.Snippet(t, testkit.SnippetSpec{})

	assert.False(t, snippet.ID().IsNil())
	assert.Equal(t, "Snippet", snippet.Title().String())
	assert.True(t, snippet.AtRoot())
	assert.Len(t, snippet.Fragments(), 1)
	assert.Equal(t, "plaintext", snippet.Fragments()[0].Language().String())
	assert.Equal(t, snippet.CreatedAt(), snippet.UpdatedAt())
}

func TestSnippet_Spec(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	ids := testkit.NewSequentialIDs()
	folderID := ids.NewFolderID()

	snippet := testkit.Snippet(t, testkit.SnippetSpec{
		Title:     "curl json",
		FolderID:  folderID,
		Fragment:  testkit.FragmentSpec{Language: "Bash", Content: "curl -d @body.json"},
		CreatedAt: created,
	})

	assert.Equal(t, "curl json", snippet.Title().String())
	assert.Equal(t, folderID, snippet.FolderID())
	assert.Equal(t, "Bash", snippet.Fragments()[0].Language().String())
	assert.Equal(t, "curl -d @body.json", snippet.Fragments()[0].Content().String())
	assert.Equal(t, created, snippet.UpdatedAt())
}
