package domain

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const fragmentsPerSnippet = 1

type Snippet struct {
	id          SnippetID
	title       value.Title
	description value.Description
	folderID    FolderID
	fragments   []Fragment
	createdAt   time.Time
	updatedAt   time.Time
}

func NewSnippet(
	id SnippetID,
	title value.Title,
	description value.Description,
	folderID FolderID,
	fragments []Fragment,
	createdAt, updatedAt time.Time,
) (Snippet, error) {
	err := errors.Join(requireID(id), requireOneFragment(fragments), requireTimestamps(createdAt, updatedAt))
	if err != nil {
		return Snippet{}, fmt.Errorf("domain.NewSnippet: %w", err)
	}

	return Snippet{
		id:          id,
		title:       title,
		description: description,
		folderID:    folderID,
		fragments:   slices.Clone(fragments),
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}, nil
}

func (s Snippet) ID() SnippetID {
	return s.id
}

func (s Snippet) Title() value.Title {
	return s.title
}

func (s Snippet) Description() value.Description {
	return s.description
}

func (s Snippet) FolderID() FolderID {
	return s.folderID
}

func (s Snippet) AtRoot() bool {
	return s.folderID.IsNil()
}

func (s Snippet) Fragments() []Fragment {
	return slices.Clone(s.fragments)
}

func (s Snippet) FirstFragment() Fragment {
	return s.fragments[0]
}

func (s Snippet) CreatedAt() time.Time {
	return s.createdAt
}

func (s Snippet) UpdatedAt() time.Time {
	return s.updatedAt
}

func (s Snippet) Edit(
	title value.Title, description value.Description, content value.Content, now time.Time,
) (Snippet, error) {
	err := requireTimestamps(s.createdAt, now)
	if err != nil {
		return Snippet{}, fmt.Errorf("domain.Snippet.Edit: %w", err)
	}

	fragment, err := s.FirstFragment().WithContent(content, now)
	if err != nil {
		return Snippet{}, fmt.Errorf("domain.Snippet.Edit: %w", err)
	}

	s.title = title
	s.description = description
	s.fragments = []Fragment{fragment}
	s.updatedAt = now

	return s, nil
}

func requireOneFragment(fragments []Fragment) error {
	if len(fragments) != fragmentsPerSnippet {
		return ErrNotOneFragment
	}

	return nil
}
