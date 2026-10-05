// Package atomicfile owns writing a file so that no reader ever sees it half
// written, so that config and state share one temporary-file-and-rename routine
// instead of each keeping its own.
package atomicfile
