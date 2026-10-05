package tui

import (
	"context"
	"io"
)

type EditorRun = func(stdin io.Reader, stdout, stderr io.Writer) (string, error)

type ExternalEditor interface {
	Open(ctx context.Context, fileName, content string) (EditorRun, error)
}
