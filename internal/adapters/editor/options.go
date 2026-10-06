package editor

import "log/slog"

type Options struct {
	Configured string
	Environ    func() []string
	LookPath   func(file string) (string, error)
	Logger     *slog.Logger
}
