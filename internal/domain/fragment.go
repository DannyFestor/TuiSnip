package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type Fragment struct {
	id        FragmentID
	language  value.Language
	content   value.Content
	createdAt time.Time
	updatedAt time.Time
}

func NewFragment(
	id FragmentID, language value.Language, content value.Content, createdAt, updatedAt time.Time,
) (Fragment, error) {
	err := errors.Join(requireID(id), requireTimestamps(createdAt, updatedAt))
	if err != nil {
		return Fragment{}, fmt.Errorf("domain.NewFragment: %w", err)
	}

	return Fragment{id: id, language: language, content: content, createdAt: createdAt, updatedAt: updatedAt}, nil
}

func (f Fragment) ID() FragmentID {
	return f.id
}

func (f Fragment) Language() value.Language {
	return f.language
}

func (f Fragment) Content() value.Content {
	return f.content
}

func (f Fragment) CreatedAt() time.Time {
	return f.createdAt
}

func (f Fragment) UpdatedAt() time.Time {
	return f.updatedAt
}
