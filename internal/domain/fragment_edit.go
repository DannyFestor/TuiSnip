package domain

import "github.com/DannyFestor/TuiSnip/internal/domain/value"

type FragmentEdit struct {
	ID       FragmentID
	Language value.Language
	Content  value.Content
}
