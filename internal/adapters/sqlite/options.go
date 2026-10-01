package sqlite

import "log/slog"

type Options struct {
	Path   string
	Logger *slog.Logger
	Clock  Clock
}
