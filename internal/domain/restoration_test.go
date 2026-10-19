package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestRestorationOf(t *testing.T) {
	t.Parallel()

	remembered := testkit.NewSequentialIDs().NewFolderID()

	tests := []struct {
		name           string
		remembered     domain.FolderID
		formerPath     value.FolderPath
		rememberedLive bool
		want           domain.Restoration
	}{
		{
			name:           "goes back into its live Folder",
			remembered:     remembered,
			rememberedLive: true,
			want:           domain.Restoration{FolderID: remembered, Fallback: domain.RestoreFallbackNone},
		},
		{
			name: "goes back to the Root it was at",
			want: domain.Restoration{FolderID: domain.FolderID{}, Fallback: domain.RestoreFallbackNone},
		},
		{
			name:       "goes to the Root when its Folder is in the Trash",
			remembered: remembered,
			want:       domain.Restoration{FolderID: domain.FolderID{}, Fallback: domain.RestoreFallbackFolderInTrash},
		},
		{
			name:       "goes to the Root when its Folder was deleted",
			formerPath: folderPath(t, "docker"),
			want:       domain.Restoration{FolderID: domain.FolderID{}, Fallback: domain.RestoreFallbackFolderDeleted},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := domain.RestorationOf(tt.remembered, tt.formerPath, tt.rememberedLive)

			assert.Equal(t, tt.want, got)
		})
	}
}
