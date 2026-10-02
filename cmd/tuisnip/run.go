package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/DannyFestor/TuiSnip/internal/adapters/xdg"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
)

const (
	exitOK      = 0
	exitFailure = 1
	exitUsage   = 2
	programName = "tuisnip"
)

var errUnexpectedArguments = errors.New("unexpected arguments")

type request struct {
	version bool
	paths   bool
}

func run(ctx context.Context, args []string, env environment) int {
	req, err := parseArgs(args, env.stderr)
	if errors.Is(err, flag.ErrHelp) {
		return exitOK
	}

	if err != nil {
		return exitUsage
	}

	if req.version {
		return exitCodeFor(env, printVersion(env.stdout))
	}

	paths, err := xdg.Resolve(env.getenv)
	if err != nil {
		return failed(env, err)
	}

	if req.paths {
		return exitCodeFor(env, printPaths(env.stdout, paths))
	}

	return exitCodeFor(env, startTUI(ctx, paths, env))
}

func parseArgs(args []string, stderr io.Writer) (request, error) {
	var req request

	flags := flag.NewFlagSet(programName, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.BoolVar(&req.version, "version", false, "print the version")
	flags.BoolVar(&req.paths, "paths", false, "print the config, data, backup, state, and log paths")

	err := flags.Parse(args)
	if err != nil {
		return request{}, fmt.Errorf("parse flags: %w", err)
	}

	if flags.NArg() > 0 {
		complain(stderr, fmt.Sprintf("%v: %v", errUnexpectedArguments, flags.Args()))
		flags.Usage()

		return request{}, errUnexpectedArguments
	}

	return req, nil
}

func startTUI(ctx context.Context, paths xdg.Paths, env environment) error {
	app, err := bootstrap.New(ctx, bootstrap.Options{
		Paths:    paths,
		Environ:  env.environ,
		LookPath: env.lookPath,
		GOOS:     env.goos,
	})
	if err != nil {
		return fmt.Errorf("start: %w", err)
	}

	err = app.Run(ctx)
	if errors.Is(err, context.Canceled) {
		err = nil
	}

	return errors.Join(err, app.Close())
}

func exitCodeFor(env environment, err error) int {
	if err == nil {
		return exitOK
	}

	return failed(env, err)
}

func failed(env environment, err error) int {
	complain(env.stderr, bootstrap.StartupMessage(err))

	return exitFailure
}

func complain(stderr io.Writer, message string) {
	//nolint:errcheck,gosec // a failed write to stderr has nowhere left to be reported
	io.WriteString(stderr, programName+": "+message+"\n")
}
