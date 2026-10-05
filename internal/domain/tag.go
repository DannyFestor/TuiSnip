package domain

import (
	"cmp"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type Tag struct {
	id        TagID
	name      value.TagName
	createdAt time.Time
	updatedAt time.Time
}

func NewTag(id TagID, name value.TagName, createdAt, updatedAt time.Time) (Tag, error) {
	err := errors.Join(requireID(id), requireTimestamps(createdAt, updatedAt))
	if err != nil {
		return Tag{}, fmt.Errorf("domain.NewTag: %w", err)
	}

	return Tag{id: id, name: name, createdAt: createdAt, updatedAt: updatedAt}, nil
}

func (t Tag) ID() TagID {
	return t.id
}

func (t Tag) Name() value.TagName {
	return t.name
}

func (t Tag) CreatedAt() time.Time {
	return t.createdAt
}

func (t Tag) UpdatedAt() time.Time {
	return t.updatedAt
}

func (t Tag) Rename(name value.TagName, now time.Time) (Tag, error) {
	err := requireTimestamps(t.createdAt, now)
	if err != nil {
		return Tag{}, fmt.Errorf("domain.Tag.Rename: %w", err)
	}

	t.name = name
	t.updatedAt = now

	return t, nil
}

func CompareTags(a, b Tag) int {
	return cmp.Or(strings.Compare(a.name.Key(), b.name.Key()), a.id.Compare(b.id))
}
