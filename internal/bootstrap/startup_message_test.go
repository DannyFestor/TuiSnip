package bootstrap_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/config"
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
)

func TestStartupMessage(t *testing.T) {
	t.Parallel()

	invalid := config.InvalidError{Path: "/home/config.toml", Problems: errors.New(`theme: "purple" is not one of`)}
	newer := sqlite.NewerSchemaError{Database: 7, Known: 5}
	other := errors.New("sqlite.Open: create data directory: not a directory")

	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "shows an invalid config without the chain", err: wrapped(invalid), want: invalid.Error()},
		{name: "shows a newer schema without the chain", err: wrapped(newer), want: newer.Error()},
		{name: "shows any other error with its chain", err: wrapped(other), want: wrapped(other).Error()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, bootstrap.StartupMessage(tt.err))
		})
	}
}

func wrapped(err error) error {
	return fmt.Errorf("bootstrap.New: open: %w", err)
}
