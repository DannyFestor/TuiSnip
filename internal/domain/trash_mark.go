package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type TrashItem interface {
	CreatedAt() time.Time
	AtRoot() bool
}

type TrashMark struct {
	deletedAt  time.Time
	formerPath value.FolderPath
}

func NewTrashMark(item TrashItem, deletedAt time.Time, formerPath value.FolderPath) (TrashMark, error) {
	err := errors.Join(
		requireStorable(deletedAt),
		requireDeletedAfterCreate(item.CreatedAt(), deletedAt),
		requireDetached(item, formerPath),
	)
	if err != nil {
		return TrashMark{}, fmt.Errorf("domain.NewTrashMark: %w", err)
	}

	return TrashMark{deletedAt: deletedAt, formerPath: formerPath}, nil
}

func (m TrashMark) DeletedAt() time.Time {
	return m.deletedAt
}

func (m TrashMark) FormerPath() value.FolderPath {
	return m.formerPath
}

func requireDeletedAfterCreate(createdAt, deletedAt time.Time) error {
	if deletedAt.Before(createdAt) {
		return ErrDeletedBeforeCreate
	}

	return nil
}

func requireDetached(item TrashItem, formerPath value.FolderPath) error {
	if !formerPath.IsEmpty() && !item.AtRoot() {
		return ErrFormerPathNotDetached
	}

	return nil
}
