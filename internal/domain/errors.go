package domain

import "errors"

var (
	ErrClipboardEmpty      = errors.New("domain: clipboard is empty")
	ErrConflict            = errors.New("domain: snippet changed elsewhere since it was loaded")
	ErrCorruptRecord       = errors.New("domain: stored record is corrupt")
	ErrEditorFailed        = errors.New("domain: external editor exited with an error")
	ErrFolderCycle         = errors.New("domain: folder cannot move into itself or its own subtree")
	ErrMissingDependency   = errors.New("domain: dependency is missing")
	ErrNilID               = errors.New("domain: id is the nil UUID")
	ErrNoClipboardTool     = errors.New("domain: no clipboard tool found")
	ErrNoEditor            = errors.New("domain: no external editor found")
	ErrNotFound            = errors.New("domain: record not found")
	ErrNotOneFragment      = errors.New("domain: snippet must have exactly one fragment")
	ErrTagNameTaken        = errors.New("domain: another tag already has this name, ignoring case")
	ErrTimestampOutOfRange = errors.New("domain: timestamp is not storable as Unix nanoseconds after the epoch")
	ErrUpdatedBeforeCreate = errors.New("domain: updated time is before created time")
)
