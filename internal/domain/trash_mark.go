package domain

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type TrashItem interface {
	CreatedAt() time.Time
	AtRoot() bool
}

type TrashMark struct {
	deletedAt  time.Time
	formerPath []value.FolderName
}

func NewTrashMark(item TrashItem, deletedAt time.Time, formerPath []value.FolderName) (TrashMark, error) {
	err := errors.Join(
		requireStorable(deletedAt),
		requireDeletedAfterCreate(item.CreatedAt(), deletedAt),
		requireDetached(item, formerPath),
	)
	if err != nil {
		return TrashMark{}, fmt.Errorf("domain.NewTrashMark: %w", err)
	}

	return TrashMark{deletedAt: deletedAt, formerPath: pathCopy(formerPath)}, nil
}

func (m TrashMark) DeletedAt() time.Time {
	return m.deletedAt
}

func (m TrashMark) FormerPath() []value.FolderName {
	return pathCopy(m.formerPath)
}

func requireDeletedAfterCreate(createdAt, deletedAt time.Time) error {
	if deletedAt.Before(createdAt) {
		return ErrDeletedBeforeCreate
	}

	return nil
}

func pathCopy(path []value.FolderName) []value.FolderName {
	if len(path) == 0 {
		return nil
	}

	return slices.Clone(path)
}

func requireDetached(item TrashItem, formerPath []value.FolderName) error {
	if len(formerPath) > 0 && !item.AtRoot() {
		return ErrFormerPathNotDetached
	}

	return nil
}
