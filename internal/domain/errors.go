package domain

import "errors"

var ErrCorruptRecord = errors.New("domain: stored record is corrupt")
