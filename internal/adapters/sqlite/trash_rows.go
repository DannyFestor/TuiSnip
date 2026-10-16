package sqlite

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqltype"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type trashEntry[E any] struct {
	entity  E
	mark    domain.TrashMark
	trashed bool
}

func entryFromColumns[E domain.TrashItem](
	entity E, deletedAt *sqltype.Timestamp, formerPath string,
) (trashEntry[E], error) {
	if deletedAt == nil {
		return trashEntry[E]{entity: entity, mark: domain.TrashMark{}, trashed: false}, nil
	}

	names, err := formerPathFromColumn(formerPath)
	if err != nil {
		return trashEntry[E]{}, corrupt(err)
	}

	mark, err := domain.NewTrashMark(entity, time.Time(*deletedAt), names)
	if err != nil {
		return trashEntry[E]{}, corrupt(err)
	}

	return trashEntry[E]{entity: entity, mark: mark, trashed: true}, nil
}

func formerPathFromColumn(column string) (value.FolderPath, error) {
	var raw []string

	err := json.Unmarshal([]byte(column), &raw)
	if err != nil {
		return value.FolderPath{}, fmt.Errorf("decode former path: %w", err)
	}

	path, err := value.NewFolderPath(raw)
	if err != nil {
		return value.FolderPath{}, fmt.Errorf("former path: %w", err)
	}

	return path, nil
}

func formerPathColumn(mark domain.TrashMark) (string, error) {
	encoded, err := json.Marshal(mark.FormerPath().Strings())
	if err != nil {
		return "", fmt.Errorf("encode former path: %w", err)
	}

	return string(encoded), nil
}

func folderTrashMarkParams(id domain.FolderID, mark domain.TrashMark) (sqlcgen.SetFolderTrashMarkParams, error) {
	formerPath, err := formerPathColumn(mark)
	if err != nil {
		return sqlcgen.SetFolderTrashMarkParams{}, err
	}

	return sqlcgen.SetFolderTrashMarkParams{
		DeletedAt:        new(sqltype.Timestamp(mark.DeletedAt())),
		FormerParentPath: formerPath,
		ID:               columnID(id),
	}, nil
}

func snippetTrashMarkParams(id domain.SnippetID, mark domain.TrashMark) (sqlcgen.SetSnippetTrashMarkParams, error) {
	formerPath, err := formerPathColumn(mark)
	if err != nil {
		return sqlcgen.SetSnippetTrashMarkParams{}, err
	}

	return sqlcgen.SetSnippetTrashMarkParams{
		DeletedAt:        new(sqltype.Timestamp(mark.DeletedAt())),
		FormerFolderPath: formerPath,
		ID:               columnID(id),
	}, nil
}

func entitiesIn[E any](entries []trashEntry[E]) []E {
	entities := make([]E, 0, len(entries))
	for _, entry := range entries {
		entities = append(entities, entry.entity)
	}

	return entities
}

func marksIn[E any, I comparable](entries []trashEntry[E], id func(entity E) I) map[I]domain.TrashMark {
	marks := make(map[I]domain.TrashMark)

	for _, entry := range entries {
		if entry.trashed {
			marks[id(entry.entity)] = entry.mark
		}
	}

	return marks
}
