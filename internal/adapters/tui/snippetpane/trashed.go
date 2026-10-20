package snippetpane

import "time"

type Trashed struct {
	Place     string
	DeletedAt time.Time
}

func (t Trashed) isSet() bool {
	return !t.DeletedAt.IsZero()
}
