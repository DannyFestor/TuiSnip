package sqlite

import "errors"

var ErrLockTimeout = errors.New("sqlite: timed out waiting for another TuiSnip to finish starting")
