package sqlite

import "time"

type Clock interface {
	Now() time.Time
}
