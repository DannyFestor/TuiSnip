package main

import (
	"flag"
	"fmt"
	"io"
	"slices"
	"strings"
)

type variants []subcommand

func variantsNamed(name string) variants {
	return slices.DeleteFunc(subcommands(), func(command subcommand) bool { return command.name != name })
}

func (v variants) parse(args []string, stderr io.Writer) (invocation, error) {
	name := v[0].name
	flags := flag.NewFlagSet(programName+" "+name, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() { writeStderr(stderr, v.usage()) }
	from := v.fromFlag(flags)

	err := flags.Parse(args)
	if err != nil {
		return invocation{}, fmt.Errorf("parse %s flags: %w", name, err)
	}

	chosen, err := v.from(*from)
	if err == nil {
		err = chosen.checkOperands(flags.Args())
	}

	if err != nil {
		complain(stderr, fmt.Sprintf("%s: %v", name, err))
		flags.Usage()

		return invocation{}, err
	}

	return invocation{subcommand: chosen, operands: flags.Args()}, nil
}

func (v variants) fromFlag(flags *flag.FlagSet) *string {
	from := new(string)

	sources := v.sources()
	if len(sources) > 0 {
		flags.StringVar(from, "from", "", "the app that wrote the file: "+strings.Join(sources, ", "))
	}

	return from
}

func (v variants) from(source string) (subcommand, error) {
	index := slices.IndexFunc(v, func(command subcommand) bool { return command.from == source })
	if index < 0 {
		return subcommand{}, fmt.Errorf("%w: %s", errUnknownSource, source)
	}

	return v[index], nil
}

func (v variants) sources() []string {
	sources := make([]string, 0, len(v))

	for _, command := range v {
		if command.from != "" {
			sources = append(sources, command.from)
		}
	}

	return sources
}

func (v variants) usage() string {
	var usage strings.Builder

	for _, command := range v {
		usage.WriteString("usage: " + programName + " " + command.synopsis() + "\n")
	}

	return usage.String()
}
