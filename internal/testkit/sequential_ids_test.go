package testkit_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestSequentialIDs_New(t *testing.T) {
	t.Parallel()

	ids := testkit.NewSequentialIDs()

	first := ids.NewSnippetID().String()
	second := ids.NewFragmentID().String()
	third := ids.NewFolderID().String()

	assert.Less(t, first, second)
	assert.Less(t, second, third)
}

func TestSequentialIDs_Repeatable(t *testing.T) {
	t.Parallel()

	first := testkit.NewSequentialIDs().NewSnippetID()
	again := testkit.NewSequentialIDs().NewSnippetID()

	assert.Equal(t, first, again)
}

func TestSequentialIDs_VersionAndVariant(t *testing.T) {
	t.Parallel()

	id := testkit.NewSequentialIDs().NewSnippetID().String()

	assert.Equal(t, "7", id[14:15], "version nibble")
	assert.Contains(t, "89ab", strings.ToLower(id[19:20]), "RFC 9562 variant nibble")
	assert.False(t, testkit.NewSequentialIDs().NewFolderID().IsNil())
}
