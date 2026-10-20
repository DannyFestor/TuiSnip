package trash_test

import (
	"testing"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

type shelved struct {
	ids     *testkit.SequentialIDs
	library domain.Library
}

func newShelved(t *testing.T) *shelved {
	t.Helper()

	return &shelved{
		ids: testkit.NewSequentialIDs(),
		library: domain.Library{
			Folders:  nil,
			Tags:     nil,
			Snippets: nil,
			Trash: domain.Trash{
				Folders:  map[domain.FolderID]domain.TrashMark{},
				Snippets: map[domain.SnippetID]domain.TrashMark{},
			},
		},
	}
}

func (s *shelved) folder(t *testing.T, name string, parentID domain.FolderID) domain.Folder {
	t.Helper()

	folder := testkit.Folder(t, testkit.FolderSpec{
		ID: s.ids.NewFolderID(), Name: name, ParentID: parentID, CreatedAt: createdAt(),
	})
	s.library.Folders = append(s.library.Folders, folder)

	return folder
}

func (s *shelved) snippet(t *testing.T, title string, folderID domain.FolderID) domain.Snippet {
	t.Helper()

	snippet := testkit.Snippet(t, testkit.SnippetSpec{
		ID:        s.ids.NewSnippetID(),
		Title:     title,
		FolderID:  folderID,
		Fragments: []testkit.FragmentSpec{{ID: s.ids.NewFragmentID(), CreatedAt: createdAt()}},
		CreatedAt: createdAt(),
	})
	s.library.Snippets = append(s.library.Snippets, snippet)

	return snippet
}

func (s *shelved) trashFolder(t *testing.T, folder domain.Folder, at time.Time) {
	t.Helper()

	s.library.Trash.Folders[folder.ID()] = testkit.TrashMark(t, testkit.TrashMarkSpec{
		Item: folder, DeletedAt: at, FormerPath: nil,
	})
}

func (s *shelved) trashSnippet(t *testing.T, snippet domain.Snippet, at time.Time, formerPath []string) {
	t.Helper()

	s.library.Trash.Snippets[snippet.ID()] = testkit.TrashMark(t, testkit.TrashMarkSpec{
		Item: snippet, DeletedAt: at, FormerPath: formerPath,
	})
}

func createdAt() time.Time {
	return time.Date(2026, time.September, 1, 9, 0, 0, 0, time.UTC)
}

func deletedAt() time.Time {
	return time.Date(2026, time.October, 14, 9, 0, 0, 0, time.UTC)
}
