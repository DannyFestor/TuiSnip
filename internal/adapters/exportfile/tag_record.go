package exportfile

import (
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type tagRecord struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func tagRecordOf(tag domain.Tag) tagRecord {
	return tagRecord{
		ID:        tag.ID().String(),
		Name:      tag.Name().String(),
		CreatedAt: timestampOf(tag.CreatedAt()),
		UpdatedAt: timestampOf(tag.UpdatedAt()),
	}
}

func tagIDOf(tag domain.Tag) string {
	return tag.ID().String()
}

func (r tagRecord) tag() (domain.Tag, error) {
	id, idErr := parseID[domain.TagID](r.ID)
	name, nameErr := value.NewTagName(r.Name)
	stamped, datesErr := parseDates(r.CreatedAt, r.UpdatedAt)

	err := errors.Join(idErr, nameErr, datesErr)
	if err != nil {
		return domain.Tag{}, r.named(err)
	}

	tag, err := domain.NewTag(id, name, stamped.created, stamped.updated)
	if err != nil {
		return domain.Tag{}, r.named(err)
	}

	return tag, nil
}

func (r tagRecord) named(err error) error {
	return fmt.Errorf("tag %s: %w", r.ID, err)
}
