package sqlite

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqltype"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

func insertTagParams(tag domain.Tag) sqlcgen.InsertTagParams {
	return sqlcgen.InsertTagParams{
		ID:        columnID(tag.ID()),
		Name:      tag.Name().String(),
		NameKey:   tag.Name().Key(),
		CreatedAt: sqltype.Timestamp(tag.CreatedAt()),
		UpdatedAt: sqltype.Timestamp(tag.UpdatedAt()),
	}
}

func updateTagParams(tag domain.Tag) sqlcgen.UpdateTagParams {
	return sqlcgen.UpdateTagParams{
		Name:      tag.Name().String(),
		NameKey:   tag.Name().Key(),
		UpdatedAt: sqltype.Timestamp(tag.UpdatedAt()),
		ID:        columnID(tag.ID()),
	}
}

func otherTagsWithNameKeyParams(tag domain.Tag) sqlcgen.ListOtherTagsWithNameKeyParams {
	return sqlcgen.ListOtherTagsWithNameKeyParams{NameKey: tag.Name().Key(), ID: columnID(tag.ID())}
}

func moveSnippetTagsParams(from, to domain.TagID) sqlcgen.MoveSnippetTagsParams {
	return sqlcgen.MoveSnippetTagsParams{ToTagID: columnID(to), FromTagID: columnID(from)}
}
