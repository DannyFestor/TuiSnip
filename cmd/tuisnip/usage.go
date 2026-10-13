package main

import (
	"flag"
	"fmt"
	"strings"
)

func printUsage(flags *flag.FlagSet) {
	writeStderr(flags.Output(), usageHeader(subcommands()))
	flags.PrintDefaults()
}

func usageHeader(commands []subcommand) string {
	width := 0
	for _, command := range commands {
		width = max(width, len(command.synopsis()))
	}

	var header strings.Builder

	header.WriteString("usage: " + programName + " [flags]\n")
	header.WriteString("       " + programName + " <command> [arguments]\n\ncommands:\n")

	for _, command := range commands {
		fmt.Fprintf(&header, "  %-*s  %s\n", width, command.synopsis(), command.summary)
	}

	header.WriteString("\nflags:\n")

	return header.String()
}
