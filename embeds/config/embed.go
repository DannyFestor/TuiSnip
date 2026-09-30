package config

import (
	_ "embed"
	"slices"
)

//go:embed default.toml
var defaultTOML []byte

func Default() []byte {
	return slices.Clone(defaultTOML)
}
