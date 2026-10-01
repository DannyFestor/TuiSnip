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
	name string
	path string
	args []string
	env  environment
}

func (t tool) run(ctx context.Context, text string) error {
	ctx, cancel := context.WithTimeout(ctx, toolTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, t.path, t.args...) //nolint:gosec // G204: a fixed tool name resolved on PATH
	cmd.Stdin = strings.NewReader(text)
	cmd.Env = t.env

	err := cmd.Run()
	if err == nil {
		return nil
	}

	if ctx.Err() != nil {
		return fmt.Errorf("%s: %w: %w", t.name, ctx.Err(), err)
	}

	return fmt.Errorf("%s: %w", t.name, err)
}
