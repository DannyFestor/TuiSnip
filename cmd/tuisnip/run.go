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

var errFlagWithCommand = errors.New("-version and -paths take no command")

type request struct {
	version bool
	paths   bool
	start   func(ctx context.Context, paths xdg.Paths, env environment) error
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

	return exitCodeFor(env, req.start(ctx, paths, env))
}

func parseArgs(args []string, stderr io.Writer) (request, error) {
	req := request{version: false, paths: false, start: startTUI}

	flags := flag.NewFlagSet(programName, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() { printUsage(flags) }
	flags.BoolVar(&req.version, "version", false, "print the version")
	flags.BoolVar(&req.paths, "paths", false, "print the config, data, backup, state, and log paths")

	err := flags.Parse(args)
	if err != nil {
		return request{}, fmt.Errorf("parse flags: %w", err)
	}

	if flags.NArg() == 0 {
		return req, nil
	}

	if req.version || req.paths {
		complain(stderr, errFlagWithCommand.Error())
		flags.Usage()

		return request{}, errFlagWithCommand
	}

	called, err := parseInvocation(flags.Args(), stderr, flags.Usage)
	if err != nil {
		return request{}, err
	}

	req.start = called.run

	return req, nil
}

func startTUI(ctx context.Context, paths xdg.Paths, env environment) error {
	return withApp(ctx, paths, env, func(app *bootstrap.App) error {
		return unlessCancelled(app.Run(ctx))
	})
}

func withApp(ctx context.Context, paths xdg.Paths, env environment, use func(app *bootstrap.App) error) error {
	app, err := bootstrap.New(ctx, bootstrap.Options{
		Paths:    paths,
		Environ:  env.environ,
		LookPath: env.lookPath,
		GOOS:     env.goos,
	})
	if err != nil {
		return fmt.Errorf("start: %w", err)
	}

	return errors.Join(use(app), app.Close())
}

func unlessCancelled(err error) error {
	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
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
	writeStderr(stderr, programName+": "+message+"\n")
}

func writeStderr(stderr io.Writer, text string) {
	//nolint:errcheck,gosec // a failed write to stderr has nowhere left to be reported
	io.WriteString(stderr, text)
}
