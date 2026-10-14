package snippetslab

import (
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type tagRecord struct {
	UUID  string `json:"uuid"`
	Title string `json:"title"`
}

func (c *conversion) tag(record tagRecord) (domain.Tag, error) {
	name, err := value.NewTagName(record.Title)
	if err != nil {
		return domain.Tag{}, record.named(err)
	}

	tag, err := domain.NewTag(c.ids.NewTagID(), name, c.now, c.now)
	if err != nil {
		return domain.Tag{}, record.named(err)
	}

	c.tagsByUUID[record.UUID] = tag

	return tag, nil
}

func (c *conversion) carriedTags(uuids []string) []domain.Tag {
	tags := make([]domain.Tag, 0, len(uuids))

	for _, uuid := range uuids {
		tag, known := c.tagsByUUID[uuid]
		if known {
			tags = append(tags, tag)
		}
	}

	return tags
}

func (r tagRecord) named(err error) error {
	return fmt.Errorf("tag %s: %w", r.UUID, err)
}
