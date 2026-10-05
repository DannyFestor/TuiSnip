package bootstrap

import (
	"fmt"
	"log/slog"

	"github.com/DannyFestor/TuiSnip/internal/adapters/clipboard"
	"github.com/DannyFestor/TuiSnip/internal/adapters/config"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
)

func clipboardOptionsFrom(options Options, logger *slog.Logger) clipboard.Options {
	return clipboard.Options{
		Environ:  options.Environ,
		LookPath: options.LookPath,
		GOOS:     options.GOOS,
		Logger:   logger,
	}
}

func newCopy(settings config.Copy, clipboardOptions clipboard.Options, finder snippet.Finder) (*snippet.Copy, error) {
	switch settings.Clipboard {
	case config.ClipboardBackendAuto:
		return copyThrough(settings, finder, clipboard.NewAuto(clipboardOptions))
	case config.ClipboardBackendNative:
		return copyThrough(settings, finder, clipboard.NewNative(clipboardOptions))
	case config.ClipboardBackendOsc52:
		return copyThrough(settings, finder, clipboard.NewOSC52())
	}

	return nil, fmt.Errorf("%w: %q", errUnknownClipboardBackend, settings.Clipboard)
}

func copyThrough(settings config.Copy, finder snippet.Finder, copier snippet.Copier) (*snippet.Copy, error) {
	build := snippet.NewCopy
	if settings.TrimTrailingNewline {
		build = snippet.NewCopyTrimmingTrailingNewline
	}

	action, err := build(finder, copier)
	if err != nil {
		return nil, fmt.Errorf("build Copy: %w", err)
	}

	return action, nil
}
