package bootstrap

import (
	"errors"

	"github.com/DannyFestor/TuiSnip/internal/adapters/config"
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite"
)

func StartupMessage(err error) string {
	if invalid, ok := errors.AsType[config.InvalidError](err); ok {
		return invalid.Error()
	}

	if newer, ok := errors.AsType[sqlite.NewerSchemaError](err); ok {
		return newer.Error()
	}

	return err.Error()
}
