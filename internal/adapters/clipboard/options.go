package clipboard

import "log/slog"

type Options struct {
	Environ  func() []string
	LookPath func(file string) (string, error)
	GOOS     string
	Logger   *slog.Logger
}
