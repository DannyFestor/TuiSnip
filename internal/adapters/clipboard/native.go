package clipboard

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Native struct {
	copying foundTool
	reading foundTool
	logger  *slog.Logger
}

type foundTool struct {
	tool  tool
	found bool
}

func NewNative(options Options) *Native {
	return nativeFor(options, environment(options.Environ()))
}

func nativeFor(options Options, env environment) *Native {
	return &Native{
		copying: lookUp(options, env, copyTools()),
		reading: lookUp(options, env, readTools()),
		logger:  options.Logger,
	}
}

func lookUp(options Options, env environment, family toolFamily) foundTool {
	chosen, found := findTool(options, env, family)

	return foundTool{tool: chosen, found: found}
}

func (n *Native) Copy(ctx context.Context, text string) (domain.CopyDelivery, error) {
	if !n.copying.found {
		return "", fmt.Errorf("clipboard.Native.Copy: %w", domain.ErrNoClipboardTool)
	}

	n.logger.DebugContext(ctx, "clipboard tool run", slog.String(keyTool, n.copying.tool.name))

	err := n.copying.tool.run(ctx, text)
	if err != nil {
		return "", fmt.Errorf("clipboard.Native.Copy: %w", err)
	}

	return domain.CopyDeliveryPlaced, nil
}

func (n *Native) ReadClipboard(ctx context.Context) (string, error) {
	if !n.reading.found {
		return "", fmt.Errorf("clipboard.Native.ReadClipboard: %w", domain.ErrNoClipboardTool)
	}

	n.logger.DebugContext(ctx, "clipboard tool run", slog.String(keyTool, n.reading.tool.name))

	text, err := n.reading.tool.output(ctx)
	if err != nil {
		return "", fmt.Errorf("clipboard.Native.ReadClipboard: %w", err)
	}

	return text, nil
}
