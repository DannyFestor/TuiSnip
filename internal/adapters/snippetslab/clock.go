package snippetslab

import "time"

type Clock interface {
	Now() time.Time
}
