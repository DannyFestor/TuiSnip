package bootstrap

import "github.com/DannyFestor/TuiSnip/internal/adapters/xdg"

type Options struct {
	Paths    xdg.Paths
	Environ  func() []string
	LookPath func(file string) (string, error)
	GOOS     string
}
