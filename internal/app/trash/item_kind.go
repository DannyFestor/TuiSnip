package trash

//go:generate go-enum

// ENUM(snippet, folder).
type ItemKind string

func (k ItemKind) IsFolder() bool {
	return k == ItemKindFolder
}
