// Package binding declares every Binding the TUI knows once, with its Scope and
// hint label, so that dispatch, the status hint and key names in text all read
// the user's keys from one table. It imports nothing from tui, so every
// component can depend on it.
package binding
