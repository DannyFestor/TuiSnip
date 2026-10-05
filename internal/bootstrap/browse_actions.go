package bootstrap

import (
	"errors"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
)

type browseActions struct {
	snippetsInFolder *browse.SnippetsInFolder
	folderTree       *browse.FolderTree
	tagList          *browse.TagList
	snippetsWithTag  *browse.SnippetsWithTag
}

func newBrowseActions(repos repositories) (browseActions, error) {
	snippetsInFolder, listErr := browse.NewSnippetsInFolder(repos.snippets)
	folderTree, treeErr := browse.NewFolderTree(repos.folders, repos.snippets)
	tagList, tagListErr := browse.NewTagList(repos.tags, repos.snippets)
	snippetsWithTag, withTagErr := browse.NewSnippetsWithTag(repos.snippets)

	return browseActions{
		snippetsInFolder: snippetsInFolder,
		folderTree:       folderTree,
		tagList:          tagList,
		snippetsWithTag:  snippetsWithTag,
	}, errors.Join(listErr, treeErr, tagListErr, withTagErr)
}
