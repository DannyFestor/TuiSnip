package domain

import (
	"bytes"
	"uuid"
)

// ID is generic over the entity it identifies, so a FolderID cannot be passed where a
// SnippetID belongs, while the methods exist once.
type ID[E any] uuid.UUID

type (
	snippetEntity  struct{}
	fragmentEntity struct{}
	folderEntity   struct{}
)

type (
	SnippetID  = ID[snippetEntity]
	FragmentID = ID[fragmentEntity]
	FolderID   = ID[folderEntity]
)

func (id ID[E]) String() string {
	return uuid.UUID(id).String()
}

func (id ID[E]) Compare(other ID[E]) int {
	return bytes.Compare(id[:], other[:])
}

func (id ID[E]) IsNil() bool {
	return uuid.UUID(id) == uuid.Nil()
}

func requireID[E any](id ID[E]) error {
	if id.IsNil() {
		return ErrNilID
	}

	return nil
}
