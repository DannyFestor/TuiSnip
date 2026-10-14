package main

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/adapters/xdg"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
)

type invocation struct {
	subcommand subcommand
	operands   []string
}

func (i invocation) run(ctx context.Context, paths xdg.Paths, env environment) error {
	return withApp(ctx, paths, env, func(app *bootstrap.App) error {
		return i.subcommand.run(ctx, app, i.operands, env.stdout)
	})
}
