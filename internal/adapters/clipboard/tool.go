package clipboard

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const toolTimeout = 2 * time.Second

type tool struct {
	name        string
	path        string
	args        []string
	env         environment
	emptyMarker string
}

func (t tool) run(ctx context.Context, text string) error {
	return t.execute(ctx, func(cmd *exec.Cmd) error {
		cmd.Stdin = strings.NewReader(text)

		return cmd.Run()
	})
}

func (t tool) output(ctx context.Context) (string, error) {
	var stdout, stderr strings.Builder

	err := t.execute(ctx, func(cmd *exec.Cmd) error {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		return cmd.Run()
	})

	switch {
	case err != nil && t.reportsEmpty(stderr.String()):
		return "", nil
	case err != nil:
		return "", err
	}

	return stdout.String(), nil
}

func (t tool) reportsEmpty(complaint string) bool {
	return t.emptyMarker != nothingToComplainAbout && strings.Contains(complaint, t.emptyMarker)
}

func (t tool) execute(ctx context.Context, started func(cmd *exec.Cmd) error) error {
	ctx, cancel := context.WithTimeout(ctx, toolTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, t.path, t.args...) //nolint:gosec // G204: a fixed tool name resolved on PATH
	cmd.Env = t.env

	err := started(cmd)
	if err == nil {
		return nil
	}

	if ctx.Err() != nil {
		return fmt.Errorf("%s: %w: %w", t.name, ctx.Err(), err)
	}

	return fmt.Errorf("%s: %w", t.name, err)
}
