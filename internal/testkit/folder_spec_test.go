package testkit_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestFolder_Defaults(t *testing.T) {
	t.Parallel()

	folder := testkit.Folder(t, testkit.FolderSpec{})

	assert.False(t, folder.ID().IsNil())
	assert.Equal(t, "Folder", folder.Name().String())
	assert.True(t, folder.AtRoot())
	assert.Equal(t, "plaintext", folder.DefaultLanguage().String())
	assert.Equal(t, folder.CreatedAt(), folder.UpdatedAt())
}

func TestFolder_Spec(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	ids := testkit.NewSequentialIDs()
	parentID := ids.NewFolderID()

	folder := testkit.Folder(t, testkit.FolderSpec{
		Name:            "docker",
		ParentID:        parentID,
		DefaultLanguage: "Bash",
		CreatedAt:       created,
	})

	assert.Equal(t, "docker", folder.Name().String())
	assert.Equal(t, parentID, folder.ParentID())
	assert.Equal(t, "Bash", folder.DefaultLanguage().String())
	assert.Equal(t, created, folder.UpdatedAt())
}
