package system_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/system"
)

const (
	versionNibble = 14
	uuidV7        = "7"
)

func TestIDs_New(t *testing.T) {
	t.Parallel()

	ids := system.NewIDs()

	generated := []string{
		ids.NewSnippetID().String(),
		ids.NewSnippetID().String(),
		ids.NewFragmentID().String(),
		ids.NewFolderID().String(),
		ids.NewTagID().String(),
	}

	for _, id := range generated {
		assert.Equal(t, uuidV7, id[versionNibble:versionNibble+1], "version nibble of %s", id)
	}

	assert.Len(t, uniqueIDs(generated), len(generated))
}

func uniqueIDs(ids []string) map[string]struct{} {
	unique := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		unique[id] = struct{}{}
	}

	return unique
}
