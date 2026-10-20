package snippetslab

import (
	"errors"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type conversion struct {
	ids        IDGenerator
	now        time.Time
	folderIDs  map[string]domain.FolderID
	tagsByUUID map[string]domain.Tag
	losses     domain.ImportLosses
}

func newConversion(ids IDGenerator, now time.Time) *conversion {
	return &conversion{
		ids:        ids,
		now:        now,
		folderIDs:  map[string]domain.FolderID{},
		tagsByUUID: map[string]domain.Tag{},
		losses: domain.ImportLosses{
			LostPins:          nil,
			LostGistLinks:     nil,
			LostAttachments:   nil,
			FolderFallbacks:   nil,
			FragmentFallbacks: nil,
		},
	}
}

// Folders and Tags are converted first, because a Snippet finds them by their
// SnippetsLab UUIDs.
func (c *conversion) library(read contents) (domain.Library, error) {
	folders, foldersErr := c.folders(read.Folders, rootParent())
	tags, tagsErr := entitiesOf(read.Tags, c.tag)
	snippets, snippetsErr := entitiesOf(read.Snippets, c.snippetEntry)

	err := errors.Join(foldersErr, tagsErr, snippetsErr)
	if err != nil {
		return domain.Library{}, err
	}

	return domain.NewLibrary(liveEntries(folders), tags, snippets), nil
}

func liveEntries(folders []domain.Folder) []domain.LibraryEntry[domain.Folder] {
	entries := make([]domain.LibraryEntry[domain.Folder], 0, len(folders))
	for _, folder := range folders {
		entries = append(entries, domain.NewLiveEntry(folder))
	}

	return entries
}

func (c *conversion) recordLosses(record snippetRecord, title value.Title) {
	if record.Pinned {
		c.losses.LostPins = append(c.losses.LostPins, title)
	}

	if record.linkedToGist() {
		c.losses.LostGistLinks = append(c.losses.LostGistLinks, title)
	}

	if record.attached() {
		c.losses.LostAttachments = append(c.losses.LostAttachments, title)
	}
}
