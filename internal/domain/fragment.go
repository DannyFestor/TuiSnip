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

func (f Fragment) Duplicate(id FragmentID, now time.Time) (Fragment, error) {
	duplicate, err := NewFragment(id, f.language, f.content, now, now)
	if err != nil {
		return Fragment{}, fmt.Errorf("domain.Fragment.Duplicate: %w", err)
	}

	return duplicate, nil
}

func (f Fragment) Edit(language value.Language, content value.Content, now time.Time) (Fragment, error) {
	err := requireTimestamps(f.createdAt, now)
	if err != nil {
		return Fragment{}, fmt.Errorf("domain.Fragment.Edit: %w", err)
	}

	if language == f.language && content == f.content {
		return f, nil
	}

	f.language = language
	f.content = content
	f.updatedAt = now

	return f, nil
}
