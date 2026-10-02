package bootstrap

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite"
	"github.com/DannyFestor/TuiSnip/internal/adapters/xdg"
)

func BackupDir(paths xdg.Paths) string {
	return sqlite.BackupDir(paths.DatabaseFile)
}
