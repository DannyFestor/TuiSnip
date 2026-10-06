package tagpane

import (
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagrefusal"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	newTagCount = "0"
	takenFormat = "Tag %s already exists"
)

type naming interface {
	placed(cursor int, tags []browse.TagCount, typed string) placement
	validated(raw string) error
	requested(name string) []outcome.Outcome
}

type placement struct {
	index  int
	insert bool
	meta   string
}

type line struct {
	text string
	meta string
}

type newTag struct {
	existing []browse.TagCount
}

type renamedTag struct {
	tagID        domain.TagID
	snippetCount int
}

func (newTag) placed(_ int, tags []browse.TagCount, typed string) placement {
	return placement{index: sortedIndex(tags, typed), insert: true, meta: newTagCount}
}

func (n newTag) validated(raw string) error {
	name, err := parseName(raw)
	if err != nil {
		return err
	}

	index := slices.IndexFunc(n.existing, func(counted browse.TagCount) bool {
		return counted.Tag.Name().Key() == name.Key()
	})
	if index >= 0 {
		return takenNameError{existing: n.existing[index].Tag.Name()}
	}

	return nil
}

func (newTag) requested(name string) []outcome.Outcome {
	return []outcome.Outcome{outcome.TagCreateRequested{Input: tag.CreateInput{Name: name}}}
}

func (r renamedTag) placed(cursor int, tags []browse.TagCount, _ string) placement {
	return placement{index: cursor, insert: len(tags) == 0, meta: strconv.Itoa(r.snippetCount)}
}

func (renamedTag) validated(raw string) error {
	_, err := parseName(raw)

	return err
}

func (r renamedTag) requested(name string) []outcome.Outcome {
	return []outcome.Outcome{outcome.TagRenameRequested{Input: tag.RenameInput{TagID: r.tagID, Name: name}}}
}

func (p placement) into(lines []line, field string) []line {
	shown := line{text: TagPrefix + field, meta: p.meta}
	if p.insert {
		return slices.Insert(lines, p.index, shown)
	}

	lines[p.index] = shown

	return lines
}

func linesOf(tags []browse.TagCount) []line {
	lines := make([]line, 0, len(tags))
	for _, counted := range tags {
		lines = append(lines, line{
			text: TagPrefix + counted.Tag.Name().String(),
			meta: strconv.Itoa(counted.SnippetCount),
		})
	}

	return lines
}

func sortedIndex(tags []browse.TagCount, typed string) int {
	key := value.TagNameKey(typed)

	index := slices.IndexFunc(tags, func(counted browse.TagCount) bool {
		return counted.Tag.Name().Key() > key
	})
	if index < 0 {
		return len(tags)
	}

	return index
}

func parseName(raw string) (value.TagName, error) {
	name, err := value.NewTagName(raw)
	if err != nil {
		return value.TagName{}, fmt.Errorf("tagpane: %w", err)
	}

	return name, nil
}

func refusalText(err error) string {
	if taken, ok := errors.AsType[takenNameError](err); ok {
		return fmt.Sprintf(takenFormat, taken.existing.String())
	}

	return tagrefusal.Text(err)
}
