package config

import (
	"fmt"

	"github.com/BurntSushi/toml"

	defaultconfig "github.com/DannyFestor/TuiSnip/embeds/config"
)

type rawBindings = map[string]map[string][]string

type rawConfig struct {
	Editor    string      `toml:"editor"`
	Theme     string      `toml:"theme"`
	Languages []string    `toml:"languages"`
	Mouse     bool        `toml:"mouse"`
	Copy      rawCopy     `toml:"copy"`
	Bindings  rawBindings `toml:"bindings"`
}

type rawCopy struct {
	Clipboard           string `toml:"clipboard"`
	TrimTrailingNewline bool   `toml:"trim_trailing_newline"`
	QuitAfter           bool   `toml:"quit_after"`
}

func decodeDefaults() (rawConfig, error) {
	var defaults rawConfig

	_, err := decodeInto(defaultconfig.Default(), &defaults)
	if err != nil {
		return rawConfig{}, fmt.Errorf("decode embedded defaults: %w", err)
	}

	return defaults, nil
}

func decodeInto(data []byte, into *rawConfig) ([]string, error) {
	metadata, err := toml.Decode(string(data), into)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	undecoded := make([]string, 0, len(metadata.Undecoded()))
	for _, key := range metadata.Undecoded() {
		undecoded = append(undecoded, key.String())
	}

	return undecoded, nil
}

func (r rawConfig) withoutBindings() rawConfig {
	r.Bindings = nil

	return r
}
