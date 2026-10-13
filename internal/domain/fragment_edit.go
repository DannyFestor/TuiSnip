package domain

import "github.com/DannyFestor/TuiSnip/internal/domain/value"

type FragmentEdit struct {
	ID       FragmentID
	Title    value.FragmentTitle
	Language value.Language
	Content  value.Content
}

func (e FragmentEdit) adds() bool {
	return e.ID.IsNil()
}
