package foldertree

import (
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	RootSnippetCount   = 2
	DockerSnippetCount = 3
	GoSnippetCount     = 2
	TestsSnippetCount  = 1

	dockerID = "0194c3a0-f01d-7000-8000-000000000001"
	goID     = "0194c3a0-f01d-7000-8000-000000000002"
	testsID  = "0194c3a0-f01d-7000-8000-000000000003"
)

type Sample struct {
	Tree   browse.Tree
	Docker domain.Folder
	Go     domain.Folder
	Tests  domain.Folder
}

func New(t *testing.T) Sample {
	t.Helper()

	docker := folder(t, dockerID, "docker", domain.FolderID{})
	golang := folder(t, goID, "go", domain.FolderID{})
	tests := folder(t, testsID, "testing", golang.ID())

	return Sample{
		Tree: browse.Tree{
			RootSnippetCount: RootSnippetCount,
			Folders: []browse.FolderNode{
				{Folder: docker, SnippetCount: DockerSnippetCount, Children: nil},
				{Folder: golang, SnippetCount: GoSnippetCount, Children: []browse.FolderNode{
					{Folder: tests, SnippetCount: TestsSnippetCount, Children: nil},
				}},
			},
		},
		Docker: docker,
		Go:     golang,
		Tests:  tests,
	}
}

func folder(t *testing.T, id, name string, parentID domain.FolderID) domain.Folder {
	t.Helper()

	folderName, err := value.NewFolderName(name)
	require.NoError(t, err)

	created := time.Date(2026, time.September, 1, 9, 0, 0, 0, time.UTC)
	built, err := domain.NewFolder(
		domain.FolderID(uuid.MustParse(id)), folderName, parentID, value.PlainText(), created, created,
	)
	require.NoError(t, err)

	return built
}
