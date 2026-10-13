package exportfile

import "time"

const timestampLayout = "2006-01-02T15:04:05.000000000Z07:00"

func timestampOf(at time.Time) string {
	return at.UTC().Format(timestampLayout)
}
