package tui

import "time"

type Clock interface {
	Now() time.Time
}
