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
