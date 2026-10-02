package main

import (
	"fmt"
	"io"
	"runtime/debug"
)

const unknownVersion = "(unknown)"

func printVersion(stdout io.Writer) error {
	_, err := fmt.Fprintln(stdout, programName, version())
	if err != nil {
		return fmt.Errorf("print version: %w", err)
	}

	return nil
}

func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return unknownVersion
	}

	return info.Main.Version
}
