package clipboard

import (
	"context"
	"log/slog"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Auto struct {
	native *Native
	osc52  *OSC52
	remote bool
	logger *slog.Logger
}

func NewAuto(options Options) *Auto {
	env := environment(options.Environ())

	return &Auto{
		native: nativeFor(options, env),
		osc52:  NewOSC52(),
		remote: isRemoteSession(env),
		logger: options.Logger,
	}
}

func (a *Auto) Copy(ctx context.Context, text string) (domain.CopyDelivery, error) {
	if a.remote {
		// A platform tool would write the remote host's clipboard, not the user's.
		return a.osc52.Copy(ctx, text)
	}

	delivery, err := a.native.Copy(ctx, text)
	if err == nil {
		return delivery, nil
	}

	a.logger.WarnContext(ctx, "clipboard tool failed, sending OSC 52", slog.Any(keyError, err))

	return a.osc52.Copy(ctx, text)
}
