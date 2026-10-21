// Package editoverlay is the edit overlay: the form for a new or a stored Snippet, the save
// it asks for and the answer it waits on, and the confirmations it opens before
// throwing unsaved changes away or saving them at the Root when their Snippet or
// Folder is gone. It colours Content once typing pauses, lexing it in a command
// so a keypress never waits on chroma. It reports outcomes and never runs an
// Action.
package editoverlay
