package config

import (
	"context"
	"fmt"
	"log/slog"

	defaultconfig "github.com/DannyFestor/TuiSnip/embeds/config"
	"github.com/DannyFestor/TuiSnip/internal/adapters/atomicfile"
)

func Load(ctx context.Context, options Options) (Config, error) {
	created, err := atomicfile.CreateIfMissing(options.Path, defaultconfig.Default())
	if err != nil {
		return Config{}, fmt.Errorf("config.Load: %w", err)
	}

	if created {
		options.Logger.InfoContext(ctx, "default config written", slog.String(keyPath, options.Path))
	}

	data, err := readUserFile(options.Path)
	if err != nil {
		return Config{}, fmt.Errorf("config.Load: %w", err)
	}

	defaults, err := decodeDefaults()
	if err != nil {
		return Config{}, fmt.Errorf("config.Load: %w", err)
	}

	return layerUserFile(ctx, options, defaults, data)
}

func layerUserFile(ctx context.Context, options Options, defaults rawConfig, data []byte) (Config, error) {
	layered := defaults.withoutBindings()

	unknownKeys, err := decodeInto(data, &layered)
	if err != nil {
		return Config{}, InvalidError{Path: options.Path, Problems: err}
	}

	warnUnknownKeys(ctx, options, unknownKeys)

	cfg, err := parse(layered, defaults.Bindings)
	if err != nil {
		return Config{}, InvalidError{Path: options.Path, Problems: err}
	}

	return cfg, nil
}

func warnUnknownKeys(ctx context.Context, options Options, keys []string) {
	for _, key := range keys {
		options.Logger.WarnContext(ctx, "unknown config key ignored",
			slog.String(keyConfigKey, key), slog.String(keyPath, options.Path))
	}
}
