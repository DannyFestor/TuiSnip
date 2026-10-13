package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
)

const standardStreamOperand = "-"

var (
	errUnknownCommand     = errors.New("unknown command")
	errWrongArgumentCount = errors.New("wrong number of arguments")
	errStandardStream     = errors.New(standardStreamOperand + " is not supported")
)

type subcommand struct {
	name         string
	operandNames []string
	summary      string
	run          func(ctx context.Context, app *bootstrap.App, operands []string, stdout io.Writer) error
}

func subcommands() []subcommand {
	return []subcommand{
		{
			name:         "export",
			operandNames: []string{"file"},
			summary:      "write every Folder, Tag, and Snippet to a JSON file",
			run:          exportLibrary,
		},
	}
}

func parseInvocation(args []string, stderr io.Writer, usage func()) (invocation, error) {
	for _, command := range subcommands() {
		if command.name == args[0] {
			return command.parse(args[1:], stderr)
		}
	}

	complain(stderr, fmt.Sprintf("%v: %s", errUnknownCommand, args[0]))
	usage()

	return invocation{}, errUnknownCommand
}

func (s subcommand) synopsis() string {
	words := make([]string, 0, 1+len(s.operandNames))

	words = append(words, s.name)
	for _, operandName := range s.operandNames {
		words = append(words, "<"+operandName+">")
	}

	return strings.Join(words, " ")
}

func (s subcommand) parse(args []string, stderr io.Writer) (invocation, error) {
	flags := flag.NewFlagSet(programName+" "+s.name, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() { writeStderr(stderr, "usage: "+programName+" "+s.synopsis()+"\n") }

	err := flags.Parse(args)
	if err != nil {
		return invocation{}, fmt.Errorf("parse %s flags: %w", s.name, err)
	}

	err = s.checkOperands(flags.Args())
	if err != nil {
		complain(stderr, fmt.Sprintf("%s: %v", s.name, err))
		flags.Usage()

		return invocation{}, err
	}

	return invocation{subcommand: s, operands: flags.Args()}, nil
}

func (s subcommand) checkOperands(operands []string) error {
	if len(operands) != len(s.operandNames) {
		return errWrongArgumentCount
	}

	if slices.Contains(operands, standardStreamOperand) {
		return errStandardStream
	}

	return nil
}
