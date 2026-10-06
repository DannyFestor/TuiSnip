package editor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	shellPath   = "/bin/sh"
	programName = "tuisnip"
	passFile    = ` "$@"`
)

type Run = func(stdin io.Reader, stdout, stderr io.Writer) (string, error)

type Launcher struct {
	command string
	found   bool
	environ []string
	logger  *slog.Logger
}

type terminal struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func New(options Options) *Launcher {
	environ := options.Environ()
	command, found := resolveCommand(options, environ)

	return &Launcher{command: command, found: found, environ: environ, logger: options.Logger}
}

func (l *Launcher) Open(ctx context.Context, fileName, content string) (Run, error) {
	if !l.found {
		return nil, fmt.Errorf("editor.Launcher.Open: %w", domain.ErrNoEditor)
	}

	return func(stdin io.Reader, stdout, stderr io.Writer) (string, error) {
		edited, err := l.edit(ctx, fileName, content, terminal{stdin: stdin, stdout: stdout, stderr: stderr})
		if err != nil {
			return "", fmt.Errorf("editor.Launcher.Open: %w", err)
		}

		return edited, nil
	}, nil
}

func (l *Launcher) edit(ctx context.Context, fileName, content string, attached terminal) (string, error) {
	dir, err := newScratchDir()
	if err != nil {
		return "", err
	}

	defer l.removed(ctx, dir)

	path, err := dir.written(fileName, content)
	if err != nil {
		return "", err
	}

	err = l.run(ctx, path, attached)
	if err != nil {
		return "", err
	}

	return readBack(path)
}

func (l *Launcher) run(ctx context.Context, path string, attached terminal) error {
	l.logger.DebugContext(ctx, "external editor run", slog.String(keyCommand, l.command))

	//nolint:gosec // G204: the user's own editor command, run the way git runs core.editor
	cmd := exec.CommandContext(ctx, shellPath, "-c", l.command+passFile, programName, path)
	cmd.Env = l.environ
	cmd.Stdin = attached.stdin
	cmd.Stdout = attached.stdout
	cmd.Stderr = attached.stderr

	return classified(cmd.Run())
}

func (l *Launcher) removed(ctx context.Context, dir scratchDir) {
	err := dir.remove()
	if err != nil {
		l.logger.WarnContext(ctx, "scratch dir left behind", slog.Any(keyError, err))
	}
}

func classified(err error) error {
	if err == nil {
		return nil
	}

	if exitErr, exited := errors.AsType[*exec.ExitError](err); exited {
		return fmt.Errorf("%w: %w", domain.ErrEditorFailed, exitErr)
	}

	return domain.EditorStartError{Err: err}
}
