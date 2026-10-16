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
			TrashedSnippets:   0,
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
	live := liveSnippets(read.Snippets)
	snippets, snippetsErr := entitiesOf(live, c.snippet)
	c.losses.TrashedSnippets = len(read.Snippets) - len(live)

	err := errors.Join(foldersErr, tagsErr, snippetsErr)
	if err != nil {
		return domain.Library{}, err
	}

	return domain.Library{
		Folders:  folders,
		Tags:     tags,
		Snippets: snippets,
		Trash:    domain.Trash{Folders: nil, Snippets: nil},
	}, nil
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
