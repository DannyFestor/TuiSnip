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
