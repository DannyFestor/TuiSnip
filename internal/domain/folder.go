package domain

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type Folder struct {
	id              FolderID
	name            value.FolderName
	parentID        FolderID
	defaultLanguage value.Language
	createdAt       time.Time
	updatedAt       time.Time
}

func NewFolder(
	id FolderID,
	name value.FolderName,
	parentID FolderID,
	defaultLanguage value.Language,
	createdAt, updatedAt time.Time,
) (Folder, error) {
	err := errors.Join(requireID(id), requireNotOwnParent(id, parentID), requireTimestamps(createdAt, updatedAt))
	if err != nil {
		return Folder{}, fmt.Errorf("domain.NewFolder: %w", err)
	}

	return Folder{
		id:              id,
		name:            name,
		parentID:        parentID,
		defaultLanguage: defaultLanguage,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}, nil
}

func NewFolderAtRoot(id FolderID, name value.FolderName, now time.Time) (Folder, error) {
	return NewFolder(id, name, FolderID{}, value.PlainText(), now, now)
}

func NewFolderIn(parent Folder, id FolderID, name value.FolderName, now time.Time) (Folder, error) {
	return NewFolder(id, name, parent.id, parent.defaultLanguage, now, now)
}

func (f Folder) ID() FolderID {
	return f.id
}

func (f Folder) Name() value.FolderName {
	return f.name
}

func (f Folder) ParentID() FolderID {
	return f.parentID
}

func (f Folder) AtRoot() bool {
	return f.parentID.IsNil()
}

func (f Folder) DefaultLanguage() value.Language {
	return f.defaultLanguage
}

func (f Folder) CreatedAt() time.Time {
	return f.createdAt
}

func (f Folder) UpdatedAt() time.Time {
	return f.updatedAt
}

func (f Folder) Rename(name value.FolderName, now time.Time) (Folder, error) {
	err := requireTimestamps(f.createdAt, now)
	if err != nil {
		return Folder{}, fmt.Errorf("domain.Folder.Rename: %w", err)
	}

	f.name = name
	f.updatedAt = now

	return f, nil
}

func (f Folder) WithDefaultLanguage(language value.Language, now time.Time) (Folder, error) {
	err := requireTimestamps(f.createdAt, now)
	if err != nil {
		return Folder{}, fmt.Errorf("domain.Folder.WithDefaultLanguage: %w", err)
	}

	f.defaultLanguage = language
	f.updatedAt = now

	return f, nil
}

func (f Folder) MoveUnder(parentID FolderID, descendantIDs []FolderID) (Folder, error) {
	if parentID == f.id || slices.Contains(descendantIDs, parentID) {
		return Folder{}, fmt.Errorf("domain.Folder.MoveUnder: %w", ErrFolderCycle)
	}

	f.parentID = parentID

	return f, nil
}

func (f Folder) MoveToRoot() Folder {
	f.parentID = FolderID{}

	return f
}

func CompareFolders(a, b Folder) int {
	return cmp.Or(a.name.CompareIgnoringCase(b.name), a.id.Compare(b.id))
}

func requireNotOwnParent(id, parentID FolderID) error {
	if id == parentID {
		return ErrFolderCycle
	}

	return nil
}
