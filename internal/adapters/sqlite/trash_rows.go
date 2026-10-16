package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqltype"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type trashColumns struct {
	deletedAt  *sqltype.Timestamp
	formerPath string
}

func entryFromColumns[E domain.TrashItem](
	entity E, deletedAt *sqltype.Timestamp, formerPath string,
) (domain.LibraryEntry[E], error) {
	path, err := formerPathFromColumn(formerPath)
	if err != nil {
		return domain.LibraryEntry[E]{}, corrupt(err)
	}

	entry, err := domain.NewLibraryEntry(entity, timeOrZero(deletedAt), path)
	if err != nil {
		return domain.LibraryEntry[E]{}, corrupt(err)
	}

	return entry, nil
}

func timeOrZero(column *sqltype.Timestamp) time.Time {
	if column == nil {
		return time.Time{}
	}

	return time.Time(*column)
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

func trashColumnsOf(mark domain.TrashMark) (trashColumns, error) {
	encoded, err := json.Marshal(mark.FormerPath().Strings())
	if err != nil {
		return trashColumns{}, fmt.Errorf("encode former path: %w", err)
	}

	return trashColumns{deletedAt: new(sqltype.Timestamp(mark.DeletedAt())), formerPath: string(encoded)}, nil
}

func folderTrashMarkParams(id domain.FolderID, columns trashColumns) sqlcgen.SetFolderTrashMarkParams {
	return sqlcgen.SetFolderTrashMarkParams{
		DeletedAt:        columns.deletedAt,
		FormerParentPath: columns.formerPath,
		ID:               columnID(id),
	}
}

func snippetTrashMarkParams(id domain.SnippetID, columns trashColumns) sqlcgen.SetSnippetTrashMarkParams {
	return sqlcgen.SetSnippetTrashMarkParams{
		DeletedAt:        columns.deletedAt,
		FormerFolderPath: columns.formerPath,
		ID:               columnID(id),
	}
}

func markTrashed[I comparable, P any](
	ctx context.Context,
	marks map[I]domain.TrashMark,
	paramsOf func(id I, columns trashColumns) P,
	setMark func(ctx context.Context, params P) error,
) error {
	for id, mark := range marks {
		columns, err := trashColumnsOf(mark)
		if err != nil {
			return err
		}

		err = setMark(ctx, paramsOf(id, columns))
		if err != nil {
			return fmt.Errorf("set trash mark: %w", err)
		}
	}

	return nil
}
