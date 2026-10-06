package sqlite

import (
	"time"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqltype"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

func insertSnippetParams(snippet domain.Snippet) sqlcgen.InsertSnippetParams {
	return sqlcgen.InsertSnippetParams{
		ID:          columnID(snippet.ID()),
		FolderID:    folderColumn(snippet.FolderID()),
		Title:       snippet.Title().String(),
		Description: snippet.Description().String(),
		CreatedAt:   sqltype.Timestamp(snippet.CreatedAt()),
		UpdatedAt:   sqltype.Timestamp(snippet.UpdatedAt()),
	}
}

func moveSnippetParams(snippet domain.Snippet) sqlcgen.MoveSnippetParams {
	return sqlcgen.MoveSnippetParams{
		FolderID: folderColumn(snippet.FolderID()),
		ID:       columnID(snippet.ID()),
	}
}

func updateSnippetParams(snippet domain.Snippet, loadedUpdatedAt time.Time) sqlcgen.UpdateSnippetParams {
	return sqlcgen.UpdateSnippetParams{
		Title:           snippet.Title().String(),
		Description:     snippet.Description().String(),
		UpdatedAt:       sqltype.Timestamp(snippet.UpdatedAt()),
		ID:              columnID(snippet.ID()),
		LoadedUpdatedAt: sqltype.Timestamp(loadedUpdatedAt),
	}
}

func updateFragmentParams(fragment domain.Fragment) sqlcgen.UpdateFragmentParams {
	return sqlcgen.UpdateFragmentParams{
		Language:  fragment.Language().String(),
		Content:   fragment.Content().String(),
		UpdatedAt: sqltype.Timestamp(fragment.UpdatedAt()),
		ID:        columnID(fragment.ID()),
	}
}

func insertSnippetTagParams(snippetID domain.SnippetID, tag domain.Tag) sqlcgen.InsertSnippetTagParams {
	return sqlcgen.InsertSnippetTagParams{SnippetID: columnID(snippetID), TagID: columnID(tag.ID())}
}

func insertFragmentParams(
	snippetID domain.SnippetID,
	position int,
	fragment domain.Fragment,
) sqlcgen.InsertFragmentParams {
	return sqlcgen.InsertFragmentParams{
		ID:        columnID(fragment.ID()),
		SnippetID: columnID(snippetID),
		Position:  int64(position),
		Language:  fragment.Language().String(),
		Content:   fragment.Content().String(),
		CreatedAt: sqltype.Timestamp(fragment.CreatedAt()),
		UpdatedAt: sqltype.Timestamp(fragment.UpdatedAt()),
	}
}
