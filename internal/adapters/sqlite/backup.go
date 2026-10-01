package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"
)

const (
	backupDirName        = "backups"
	backupNamePrefix     = "tuisnip-"
	backupNameSchema     = "-schema"
	backupNameExtension  = ".db"
	backupNamePattern    = backupNamePrefix + "*" + backupNameSchema + "*" + backupNameExtension
	backupTimeLayout     = "20060102T150405Z"
	backupsKept          = 3
	backupDirPermissions = 0o700
)

func BackupDir(databasePath string) string {
	return filepath.Join(filepath.Dir(withoutURIScheme(databasePath)), backupDirName)
}

func backUp(ctx context.Context, db *sql.DB, options Options, schema int64) (string, error) {
	dir := BackupDir(options.Path)

	err := os.MkdirAll(dir, backupDirPermissions)
	if err != nil {
		return "", fmt.Errorf("create backup directory: %w", err)
	}

	path := filepath.Join(dir, backupName(options.Clock.Now(), schema))

	_, err = db.ExecContext(ctx, "VACUUM INTO ?", path)
	if err != nil {
		return "", fmt.Errorf("back up database: %w", err)
	}

	return path, pruneBackups(dir)
}

func backupName(at time.Time, schema int64) string {
	return backupNamePrefix + at.UTC().Format(backupTimeLayout) +
		backupNameSchema + strconv.FormatInt(schema, 10) + backupNameExtension
}

func pruneBackups(dir string) error {
	paths, err := filepath.Glob(filepath.Join(dir, backupNamePattern))
	if err != nil {
		return fmt.Errorf("list backups: %w", err)
	}

	slices.Sort(paths)

	for _, path := range paths[:max(0, len(paths)-backupsKept)] {
		err = os.Remove(path)
		if err != nil {
			return fmt.Errorf("remove old backup: %w", err)
		}
	}

	return nil
}
