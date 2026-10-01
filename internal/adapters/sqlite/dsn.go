package sqlite

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

const (
	uriScheme = "file"

	pragmaKey         = "_pragma"
	pragmaForeignKeys = "foreign_keys(1)"
	pragmaJournalMode = "journal_mode(WAL)"
	pragmaSynchronous = "synchronous(NORMAL)"
	pragmaBusyTimeout = "busy_timeout(5000)"

	txLockKey       = "_txlock"
	txLockImmediate = "immediate"
)

func withoutURIScheme(path string) string {
	return strings.TrimPrefix(path, uriScheme+":")
}

func dataSourceName(path string) (string, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve database path: %w", err)
	}

	uri := new(url.URL)
	uri.Scheme = uriScheme
	uri.Path = absolutePath
	uri.RawQuery = connectionQuery().Encode()

	return uri.String(), nil
}

func connectionQuery() url.Values {
	query := url.Values{}
	for _, pragma := range []string{pragmaForeignKeys, pragmaJournalMode, pragmaSynchronous, pragmaBusyTimeout} {
		query.Add(pragmaKey, pragma)
	}

	query.Set(txLockKey, txLockImmediate)

	return query
}
