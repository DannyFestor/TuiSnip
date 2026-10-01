package config

import "github.com/DannyFestor/TuiSnip/internal/domain/value"

type Config struct {
	Editor    string
	Theme     Theme
	Languages []value.Language
	Mouse     bool
	Copy      Copy
	Bindings  map[Scope]map[Binding][]string
}
