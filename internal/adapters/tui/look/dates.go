package look

import "time"

const dateLayout = "2006-01-02"

func Date(at time.Time, location *time.Location) string {
	return at.In(location).Format(dateLayout)
}

func DeletedOn(at time.Time, location *time.Location) string {
	return "deleted " + Date(at, location)
}
