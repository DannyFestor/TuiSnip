// Package savegate keeps saves landing in the order they were started. Bubble
// Tea runs each command on its own goroutine, so without it a save started
// earlier could reach the file after a later one and overwrite it.
package savegate
