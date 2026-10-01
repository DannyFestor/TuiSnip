package sqlite

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqltype"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type uuidBytes interface {
	~[16]byte
}

func columnID[I uuidBytes](id I) sqltype.ID {
	return sqltype.ID(id)
}

func entityID[I uuidBytes](id sqltype.ID) I {
	return I(id)
}

func folderColumn(folderID domain.FolderID) *sqltype.ID {
	if folderID.IsNil() {
		return nil
	}

	id := columnID(folderID)

	return &id
}

func folderFromColumn(folderID *sqltype.ID) domain.FolderID {
	if folderID == nil {
		return domain.FolderID{}
	}

	return entityID[domain.FolderID](*folderID)
}
