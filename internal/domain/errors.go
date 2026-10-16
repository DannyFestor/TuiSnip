package domain

import "errors"

var (
	ErrClipboardEmpty        = errors.New("domain: clipboard is empty")
	ErrConflict              = errors.New("domain: snippet changed elsewhere since it was loaded")
	ErrCorruptRecord         = errors.New("domain: stored record is corrupt")
	ErrDeletedBeforeCreate   = errors.New("domain: deleted time is before created time")
	ErrEditorFailed          = errors.New("domain: external editor exited with an error")
	ErrFolderCycle           = errors.New("domain: folder cannot move into itself or its own subtree")
	ErrFormerPathNotDetached = errors.New(
		"domain: a former path belongs only to a Trash item detached from its deleted Folder",
	)
	ErrFormerPathOutsideTrash = errors.New("domain: a former path belongs only to a Trash item")
	ErrLibraryNotEmpty        = errors.New("domain: library already holds folders, tags, or snippets")
	ErrMissingDependency      = errors.New("domain: dependency is missing")
	ErrMissingFolder          = errors.New("domain: refers to a folder the library lacks")
	ErrMissingTag             = errors.New("domain: refers to a tag the library lacks")
	ErrNilID                  = errors.New("domain: id is the nil UUID")
	ErrNoClipboardTool        = errors.New("domain: no clipboard tool found")
	ErrNoEditor               = errors.New("domain: no external editor found")
	ErrNoFragment             = errors.New("domain: snippet must have at least one fragment")
	ErrNotFound               = errors.New("domain: record not found")
	ErrTagNameTaken           = errors.New("domain: another tag already has this name, ignoring case")
	ErrTimestampOutOfRange    = errors.New(
		"domain: timestamp is not storable as Unix nanoseconds after the epoch",
	)
	ErrUpdatedBeforeCreate = errors.New("domain: updated time is before created time")
)
