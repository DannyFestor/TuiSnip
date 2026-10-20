package snippetpane

import "time"

type Trashed struct {
	Place     string
	DeletedAt time.Time
}

func DeletedOn(at time.Time, location *time.Location) string {
	return "deleted " + at.In(location).Format(dateLayout)
}

func (t Trashed) isSet() bool {
	return !t.DeletedAt.IsZero()
}
