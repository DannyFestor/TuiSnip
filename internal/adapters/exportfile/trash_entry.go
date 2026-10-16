package exportfile

import (
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type trashEntry[E any] struct {
	entity  E
	mark    domain.TrashMark
	trashed bool
}

func entryOf[E domain.TrashItem](entity E, deletedAt *string, formerPath []string) (trashEntry[E], error) {
	if deletedAt == nil {
		return trashEntry[E]{entity: entity, mark: domain.TrashMark{}, trashed: false}, requireLive(formerPath)
	}

	deleted, dateErr := parseTimestamp(*deletedAt)
	path, pathErr := value.NewFolderPath(formerPath)

	err := errors.Join(dateErr, pathErr)
	if err != nil {
		return trashEntry[E]{}, err
	}

	mark, err := domain.NewTrashMark(entity, deleted, path)
	if err != nil {
		return trashEntry[E]{}, fmt.Errorf("trash mark: %w", err)
	}

	return trashEntry[E]{entity: entity, mark: mark, trashed: true}, nil
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

func requireLive(formerPath []string) error {
	if len(formerPath) > 0 {
		return domain.ErrFormerPathNotDetached
	}

	return nil
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
