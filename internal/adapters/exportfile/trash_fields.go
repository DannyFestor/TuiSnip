package exportfile

import (
	"errors"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type trashFields struct {
	deletedAt  time.Time
	formerPath value.FolderPath
}

func parseTrashFields(deletedAt *string, formerPath []string) (trashFields, error) {
	deleted, dateErr := parseDeletedAt(deletedAt)
	path, pathErr := value.NewFolderPath(formerPath)

	err := errors.Join(dateErr, pathErr)
	if err != nil {
		return trashFields{}, err
	}

	return trashFields{deletedAt: deleted, formerPath: path}, nil
}

func parseDeletedAt(text *string) (time.Time, error) {
	if text == nil {
		return time.Time{}, nil
	}

	return parseTimestamp(*text)
}

func deletedAtOf[I comparable](marks map[I]domain.TrashMark, id I) *string {
	mark, trashed := marks[id]
	if !trashed {
		return nil
	}

	return new(timestampOf(mark.DeletedAt()))
}

func formerPathOf[I comparable](marks map[I]domain.TrashMark, id I) []string {
	return marks[id].FormerPath().Strings()
}
