package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
)

const (
	standardStreamOperand = "-"
	fileOperand           = "file"
	importCommand         = "import"
	snippetsLabSource     = "snippetslab"
)

var (
	errUnknownCommand     = errors.New("unknown command")
	errUnknownSource      = errors.New("unknown source")
	errWrongArgumentCount = errors.New("wrong number of arguments")
	errStandardStream     = errors.New(standardStreamOperand + " is not supported")
)

type subcommand struct {
	name         string
	from         string
	operandNames []string
	summary      string
	run          func(ctx context.Context, app *bootstrap.App, operands []string, stdout io.Writer) error
}

func subcommands() []subcommand {
	return []subcommand{
		{
			name:         "export",
			from:         "",
			operandNames: []string{fileOperand},
			summary:      "write every Folder, Tag, and Snippet to a JSON file",
			run:          exportLibrary,
		},
		{
			name:         importCommand,
			from:         "",
			operandNames: []string{fileOperand},
			summary:      "add every Folder, Tag, and Snippet in an Export to an empty Library",
			run:          importLibrary,
		},
		{
			name:         importCommand,
			from:         snippetsLabSource,
			operandNames: []string{fileOperand},
			summary:      "add every live Snippet, Folder, and Tag in a SnippetsLab JSON export beside the Library",
			run:          importSnippetsLab,
		},
	}
}

func parseInvocation(args []string, stderr io.Writer, usage func()) (invocation, error) {
	named := variantsNamed(args[0])
	if len(named) == 0 {
		complain(stderr, fmt.Sprintf("%v: %s", errUnknownCommand, args[0]))
		usage()

		return invocation{}, errUnknownCommand
	}

	return named.parse(args[1:], stderr)
}

func (s subcommand) synopsis() string {
	words := []string{s.name}
	if s.from != "" {
		words = append(words, "--from", s.from)
	}

	for _, operandName := range s.operandNames {
		words = append(words, "<"+operandName+">")
	}

	return strings.Join(words, " ")
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
