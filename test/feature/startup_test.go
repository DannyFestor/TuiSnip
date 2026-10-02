//go:build feature

package feature_test

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"github.com/DannyFestor/TuiSnip/embeds/config"
	adapterconfig "github.com/DannyFestor/TuiSnip/internal/adapters/config"
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

const (
	invalidConfig      = "theme = \"purple\"\n[copy]\nclipboard = \"carrier-pigeon\"\n"
	newerSchemaVersion = 99
)

func TestStartupWritesDefaultConfig(t *testing.T) {
	t.Parallel()

	home, _ := testapp.Start(t, testapp.RecordingTool)

	written, err := os.ReadFile(home.Paths.ConfigFile)
	require.NoError(t, err)
	assert.Equal(t, config.Default(), written)
}

func TestStartupLogsTheSession(t *testing.T) {
	t.Parallel()

	home, _ := testapp.Start(t, testapp.RecordingTool)

	logged := home.Log(t)
	assert.Contains(t, logged, `msg="tuisnip started"`)
	assert.Contains(t, logged, "session=")
	assert.Contains(t, logged, "database_path="+home.Paths.DatabaseFile)
}

func TestStartupRefusesInvalidConfig(t *testing.T) {
	t.Parallel()

	home := testapp.NewHome(t)
	home.WriteConfig(t, invalidConfig)

	_, err := home.Open(t, testapp.RecordingTool)

	invalid, ok := errors.AsType[adapterconfig.InvalidError](err)
	require.True(t, ok, "want an InvalidError, got %v", err)
	assert.Contains(t, invalid.Error(), "theme")
	assert.Contains(t, invalid.Error(), "copy.clipboard")
	assert.NoFileExists(t, home.Paths.DatabaseFile)
	assert.Contains(t, home.Log(t), `msg="start-up failed"`)
}

func TestStartupRefusesNewerSchema(t *testing.T) {
	t.Parallel()

	home := testapp.NewHome(t)
	first, err := home.Open(t, testapp.RecordingTool)
	require.NoError(t, err)
	require.NoError(t, first.Close())
	markSchemaVersion(t, home.Paths.DatabaseFile, newerSchemaVersion)

	_, err = home.Open(t, testapp.RecordingTool)

	newer, ok := errors.AsType[sqlite.NewerSchemaError](err)
	require.True(t, ok, "want a NewerSchemaError, got %v", err)
	assert.Equal(t, int64(newerSchemaVersion), newer.Database)
}

func TestStartupFailsWhenDataDirIsBlocked(t *testing.T) {
	t.Parallel()

	home := testapp.NewHome(t)
	home.Block(t, filepath.Dir(home.Paths.DatabaseFile))

	_, err := home.Open(t, testapp.RecordingTool)

	require.ErrorContains(t, err, "create data directory")
}

func TestStartupFailsWhenStateDirIsBlocked(t *testing.T) {
	t.Parallel()

	home := testapp.NewHome(t)
	home.Block(t, filepath.Dir(home.Paths.LogFile))

	_, err := home.Open(t, testapp.RecordingTool)

	require.ErrorContains(t, err, "create log directory")
	assert.NoFileExists(t, home.Paths.ConfigFile)
}

func markSchemaVersion(t *testing.T, databasePath string, version int) {
	t.Helper()

	db, err := sql.Open("sqlite", databasePath)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	_, err = db.ExecContext(t.Context(), "INSERT INTO goose_db_version (version_id, is_applied) VALUES (?, 1)", version)
	require.NoError(t, err)
}
