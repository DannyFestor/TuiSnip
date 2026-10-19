package sqlite_test

import (
	"io/fs"
	"os"
	"regexp"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	queriesDir = "../../../db/queries"

	write         = "a write"
	library       = "Export reads the whole Library, the Trash included"
	emptiness     = "the Import emptiness check counts the whole Library, the Trash included"
	joinedToLive  = "joins snippet_tag to live_snippets"
	wholeTreeOnly = "the Folder cycle check guards the whole tree, the Trash included"
	trashItems    = "the Trash list, its count, restore, and Copy read the Trash itself"
)

var (
	queryHeader      = regexp.MustCompile(`(?m)^-- name: (\w+) :\w+$`)
	baseTableRead    = regexp.MustCompile(`(?i)\b(?:FROM|JOIN)\s+(?:snippets|folders|snippet_tag)\b`)
	liveSnippetsJoin = regexp.MustCompile(`(?i)\b(?:FROM|JOIN)\s+live_snippets\b`)
)

type namedQuery struct {
	name string
	body string
}

func TestQueries_baseTableAllowlist(t *testing.T) {
	t.Parallel()

	t.Run("lets only allowlisted queries name snippets, folders, or snippet_tag", func(t *testing.T) {
		t.Parallel()

		unlisted := unlistedBaseTableReads(storedQueries(t), baseTableReaders())

		assert.Empty(t, unlisted, "read live_snippets or live_folders, or allowlist the query with a reason")
	})

	t.Run("keeps no allowlist entry for a query that names no base table", func(t *testing.T) {
		t.Parallel()

		readers := baseTableReadersIn(storedQueries(t))

		for name := range baseTableReaders() {
			assert.Contains(t, readers, name)
		}
	})

	t.Run("keeps the live_snippets join in every query allowlisted for it", func(t *testing.T) {
		t.Parallel()

		allowlist := baseTableReaders()

		for _, query := range storedQueries(t) {
			if allowlist[query.name] == joinedToLive {
				assert.Regexp(t, liveSnippetsJoin, query.body, query.name)
			}
		}
	})

	t.Run("flags a new query that reads snippets directly", func(t *testing.T) {
		t.Parallel()

		source := "-- name: ListEverything :many\nSELECT id FROM snippets;\n\n" +
			"-- name: ListLiveEverything :many\nSELECT id FROM live_snippets;\n"

		unlisted := unlistedBaseTableReads(queriesIn(source), baseTableReaders())

		assert.Equal(t, []string{"ListEverything"}, unlisted)
	})
}

func baseTableReaders() map[string]string {
	return map[string]string{
		"CountFolderTrashItems":        trashItems,
		"CountSnippetTrashItems":       trashItems,
		"GetFolderTrashItem":           trashItems,
		"GetSnippetTrashItem":          trashItems,
		"ListSnippetTrashItemTags":     trashItems,
		"ListTrashSnippetTags":         trashItems,
		"ListTrashSnippets":            trashItems,
		"CountFolders":                 emptiness,
		"CountSnippets":                emptiness,
		"CountSnippetsByTag":           joinedToLive,
		"CountSnippetsWithTag":         joinedToLive,
		"DeleteSnippetTags":            write,
		"ListDescendantFolderIDs":      wholeTreeOnly,
		"ListFragmentsWithTag":         joinedToLive,
		"ListLibraryFolders":           library,
		"ListLibrarySnippetTags":       library,
		"ListLibrarySnippets":          library,
		"ListLiveSnippetTags":          joinedToLive,
		"ListSnippetTagsBySnippet":     joinedToLive,
		"ListSnippetTagsInFolder":      joinedToLive,
		"ListSnippetTagsWithTag":       joinedToLive,
		"ListSnippetsWithTagByCreated": joinedToLive,
		"ListSnippetsWithTagByTitle":   joinedToLive,
		"ListSnippetsWithTagByUpdated": joinedToLive,
		"MoveSnippetTags":              write,
	}
}

func unlistedBaseTableReads(queries []namedQuery, allowlist map[string]string) []string {
	var unlisted []string

	for _, name := range baseTableReadersIn(queries) {
		if _, listed := allowlist[name]; !listed {
			unlisted = append(unlisted, name)
		}
	}

	return unlisted
}

func baseTableReadersIn(queries []namedQuery) []string {
	var readers []string

	for _, query := range queries {
		if baseTableRead.MatchString(query.body) {
			readers = append(readers, query.name)
		}
	}

	return readers
}

func storedQueries(t *testing.T) []namedQuery {
	t.Helper()

	files := os.DirFS(queriesDir)
	names, err := fs.Glob(files, "*.sql")
	require.NoError(t, err)
	require.NotEmpty(t, names)

	perFile := make([][]namedQuery, 0, len(names))

	for _, name := range names {
		source, readErr := fs.ReadFile(files, name)
		require.NoError(t, readErr)

		perFile = append(perFile, queriesIn(string(source)))
	}

	return slices.Concat(perFile...)
}

func queriesIn(source string) []namedQuery {
	headers := queryHeader.FindAllStringSubmatchIndex(source, -1)
	queries := make([]namedQuery, 0, len(headers))

	for index, header := range headers {
		end := len(source)
		if index+1 < len(headers) {
			end = headers[index+1][0]
		}

		queries = append(queries, namedQuery{name: source[header[2]:header[3]], body: source[header[1]:end]})
	}

	return queries
}
