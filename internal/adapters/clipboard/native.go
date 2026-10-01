package clipboard

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Native struct {
	tool   tool
	found  bool
	logger *slog.Logger
}

func NewNative(options Options) *Native {
	return nativeFor(options, environment(options.Environ()))
}

func nativeFor(options Options, env environment) *Native {
	chosen, found := findTool(options, env)

	return &Native{tool: chosen, found: found, logger: options.Logger}
}

func (n *Native) Copy(ctx context.Context, text string) (domain.CopyDelivery, error) {
	if !n.found {
		return "", fmt.Errorf("clipboard.Native.Copy: %w", domain.ErrNoClipboardTool)
	}

	n.logger.DebugContext(ctx, "clipboard tool run", slog.String(keyTool, n.tool.name))

	err := n.tool.run(ctx, text)
	if err != nil {
		return "", fmt.Errorf("clipboard.Native.Copy: %w", err)
	}

	return domain.CopyDeliveryPlaced, nil
}
