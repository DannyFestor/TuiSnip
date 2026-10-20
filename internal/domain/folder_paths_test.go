package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestFolderPaths_To(t *testing.T) {
	t.Parallel()

	ids := testkit.NewSequentialIDs()
	ops := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "ops"})
	docker := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "a/b", ParentID: ops.ID()})
	detached := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "compose"})
	inDetached := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "up", ParentID: detached.ID()})
	paths := domain.NewFolderPaths([]domain.Folder{ops, docker, detached, inDetached}, domain.Trash{
		Folders: map[domain.FolderID]domain.TrashMark{
			detached.ID(): testkit.TrashMark(t, testkit.TrashMarkSpec{
				Item: detached, DeletedAt: detached.CreatedAt(), FormerPath: []string{"gone", "away"},
			}),
		},
		Snippets: map[domain.SnippetID]domain.TrashMark{},
	})

	tests := []struct {
		name     string
		folderID domain.FolderID
		want     []string
	}{
		{
			name:     "names each Folder from the Root down, the Folder itself last",
			folderID: docker.ID(),
			want:     []string{"ops", "a/b"},
		},
		{
			name:     "starts with the former path of a detached Folder at the top",
			folderID: inDetached.ID(),
			want:     []string{"gone", "away", "compose", "up"},
		},
		{name: "is empty for the Root", folderID: domain.FolderID{}, want: []string{}},
		{name: "is empty for an unknown Folder", folderID: ids.NewFolderID(), want: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, paths.To(tt.folderID))
		})
	}
}
