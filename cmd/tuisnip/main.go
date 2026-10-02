package main

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	code := run(ctx, os.Args[1:], processEnvironment())

	stop()
	os.Exit(code)
}

func processEnvironment() environment {
	return environment{
		stdout:   os.Stdout,
		stderr:   os.Stderr,
		getenv:   os.Getenv,
		environ:  os.Environ,
		lookPath: exec.LookPath,
		goos:     runtime.GOOS,
	}
}
