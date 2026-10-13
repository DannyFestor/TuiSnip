package domain

import (
	"cmp"
	"slices"
	"time"
)

type Library struct {
	Folders  []Folder
	Tags     []Tag
	Snippets []Snippet
}

func (l Library) InCreatedOrder() Library {
	return Library{
		Folders:  slices.SortedFunc(slices.Values(l.Folders), byCreated(Folder.CreatedAt, Folder.ID)),
		Tags:     slices.SortedFunc(slices.Values(l.Tags), byCreated(Tag.CreatedAt, Tag.ID)),
		Snippets: slices.SortedFunc(slices.Values(l.Snippets), byCreated(Snippet.CreatedAt, Snippet.ID)),
	}
}

func byCreated[T, E any](createdAt func(T) time.Time, id func(T) ID[E]) func(a, b T) int {
	return func(a, b T) int {
		return cmp.Or(createdAt(a).Compare(createdAt(b)), id(a).Compare(id(b)))
	}
}
