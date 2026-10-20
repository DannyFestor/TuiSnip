package domain

import (
	"fmt"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type LibraryEntry[E TrashItem] struct {
	entity  E
	mark    TrashMark
	trashed bool
}

func NewLibraryEntry[E TrashItem](
	entity E, deletedAt time.Time, formerPath value.FolderPath,
) (LibraryEntry[E], error) {
	if deletedAt.IsZero() {
		return liveEntry(entity, formerPath)
	}

	mark, err := NewTrashMark(entity, deletedAt, formerPath)
	if err != nil {
		return LibraryEntry[E]{}, err
	}

	return LibraryEntry[E]{entity: entity, mark: mark, trashed: true}, nil
}

func NewLiveEntry[E TrashItem](entity E) LibraryEntry[E] {
	return LibraryEntry[E]{
		entity:  entity,
		mark:    TrashMark{deletedAt: time.Time{}, formerPath: value.FolderPath{}},
		trashed: false,
	}
}

func liveEntry[E TrashItem](entity E, formerPath value.FolderPath) (LibraryEntry[E], error) {
	if !formerPath.IsEmpty() {
		return LibraryEntry[E]{}, fmt.Errorf("domain.NewLibraryEntry: %w", ErrFormerPathOutsideTrash)
	}

	return NewLiveEntry(entity), nil
}

func entitiesIn[E TrashItem](entries []LibraryEntry[E]) []E {
	entities := make([]E, 0, len(entries))
	for _, entry := range entries {
		entities = append(entities, entry.entity)
	}

	return entities
}

func marksIn[E TrashItem, I comparable](entries []LibraryEntry[E], id func(entity E) I) map[I]TrashMark {
	marks := make(map[I]TrashMark)

	for _, entry := range entries {
		if entry.trashed {
			marks[id(entry.entity)] = entry.mark
		}
	}

	return marks
}
