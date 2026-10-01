package domain

import "errors"

var (
	ErrCorruptRecord       = errors.New("domain: stored record is corrupt")
	ErrFolderCycle         = errors.New("domain: folder cannot move into itself or its own subtree")
	ErrNilID               = errors.New("domain: id is the nil UUID")
	ErrNotOneFragment      = errors.New("domain: snippet must have exactly one fragment")
	ErrTimestampOutOfRange = errors.New("domain: timestamp is not storable as Unix nanoseconds after the epoch")
	ErrUpdatedBeforeCreate = errors.New("domain: updated time is before created time")
)
