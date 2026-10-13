package domain

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type Snippet struct {
	id          SnippetID
	title       value.Title
	description value.Description
	folderID    FolderID
	fragments   []Fragment
	tags        []Tag
	createdAt   time.Time
	updatedAt   time.Time
}

func NewSnippet(
	id SnippetID,
	title value.Title,
	description value.Description,
	folderID FolderID,
	fragments []Fragment,
	tags []Tag,
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
		tags:        slices.SortedFunc(slices.Values(tags), CompareTags),
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

func (s Snippet) Tags() []Tag {
	return slices.Clone(s.tags)
}

func (s Snippet) CreatedAt() time.Time {
	return s.createdAt
}

func (s Snippet) UpdatedAt() time.Time {
	return s.updatedAt
}

func (s Snippet) Fragment(id FragmentID) (Fragment, error) {
	index, err := s.FragmentIndex(id)
	if err != nil {
		return Fragment{}, fmt.Errorf("domain.Snippet.Fragment: %w", err)
	}

	return s.fragments[index], nil
}

func (s Snippet) FragmentIndex(id FragmentID) (int, error) {
	index := slices.IndexFunc(s.fragments, func(fragment Fragment) bool { return fragment.ID() == id })
	if index < 0 {
		return 0, fmt.Errorf("domain.Snippet.FragmentIndex: %w", ErrNotFound)
	}

	return index, nil
}

func (s Snippet) Edit(
	title value.Title,
	description value.Description,
	edits []FragmentEdit,
	newFragmentID func() FragmentID,
	now time.Time,
) (Snippet, error) {
	err := errors.Join(requireTimestamps(s.createdAt, now), requireOneFragment(edits))
	if err != nil {
		return Snippet{}, fmt.Errorf("domain.Snippet.Edit: %w", err)
	}

	fragments, err := s.editedFragments(edits, newFragmentID, now)
	if err != nil {
		return Snippet{}, fmt.Errorf("domain.Snippet.Edit: %w", err)
	}

	s.title = title
	s.description = description
	s.fragments = fragments
	s.updatedAt = now

	return s, nil
}

func (s Snippet) Retagged(tags []Tag) Snippet {
	s.tags = slices.SortedFunc(slices.Values(tags), CompareTags)

	return s
}

func (s Snippet) Duplicate(id SnippetID, newFragmentID func() FragmentID, now time.Time) (Snippet, error) {
	fragments, err := s.duplicatedFragments(newFragmentID, now)
	if err != nil {
		return Snippet{}, fmt.Errorf("domain.Snippet.Duplicate: %w", err)
	}

	duplicate, err := NewSnippet(id, s.title, s.description, s.folderID, fragments, s.tags, now, now)
	if err != nil {
		return Snippet{}, fmt.Errorf("domain.Snippet.Duplicate: %w", err)
	}

	return duplicate, nil
}

func (s Snippet) MoveTo(folderID FolderID) Snippet {
	s.folderID = folderID

	return s
}

func (s Snippet) editedFragments(
	edits []FragmentEdit, newFragmentID func() FragmentID, now time.Time,
) ([]Fragment, error) {
	fragments := make([]Fragment, 0, len(edits))

	for _, edit := range edits {
		fragment, err := s.editedFragment(edit, newFragmentID, now)
		if err != nil {
			return nil, err
		}

		fragments = append(fragments, fragment)
	}

	return fragments, nil
}

func (s Snippet) editedFragment(edit FragmentEdit, newFragmentID func() FragmentID, now time.Time) (Fragment, error) {
	if edit.adds() {
		return NewFragment(newFragmentID(), edit.Title, edit.Language, edit.Content, now, now)
	}

	stored, err := s.Fragment(edit.ID)
	if err != nil {
		return Fragment{}, err
	}

	return stored.Edit(edit.Title, edit.Language, edit.Content, now)
}

func (s Snippet) duplicatedFragments(newFragmentID func() FragmentID, now time.Time) ([]Fragment, error) {
	fragments := make([]Fragment, 0, len(s.fragments))

	for _, fragment := range s.fragments {
		duplicate, err := fragment.Duplicate(newFragmentID(), now)
		if err != nil {
			return nil, err
		}

		fragments = append(fragments, duplicate)
	}

	return fragments, nil
}

func requireOneFragment[E Fragment | FragmentEdit](fragments []E) error {
	if len(fragments) == 0 {
		return ErrNoFragment
	}

	return nil
}
