package sqltype

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"uuid"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type ID uuid.UUID

var (
	_ driver.Valuer = ID{}
	_ sql.Scanner   = (*ID)(nil)
)

func (id ID) Value() (driver.Value, error) {
	return uuid.UUID(id).String(), nil
}

func (id *ID) Scan(src any) error {
	text, err := idText(src)
	if err != nil {
		return err
	}

	parsed, err := uuid.Parse(text)
	if err != nil {
		return fmt.Errorf("sqltype.ID.Scan: %w: %w", domain.ErrCorruptRecord, err)
	}

	*id = ID(parsed)

	return nil
}

func idText(src any) (string, error) {
	switch text := src.(type) {
	case string:
		return text, nil
	case []byte:
		return string(text), nil
	default:
		return "", fmt.Errorf("sqltype.ID.Scan: unsupported column type %T: %w", src, domain.ErrCorruptRecord)
	}
}
