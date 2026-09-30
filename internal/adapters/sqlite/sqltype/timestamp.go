package sqltype

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Timestamp time.Time

var (
	_ driver.Valuer = Timestamp{}
	_ sql.Scanner   = (*Timestamp)(nil)
)

func (t Timestamp) Value() (driver.Value, error) {
	return time.Time(t).UnixNano(), nil
}

func (t *Timestamp) Scan(src any) error {
	nanoseconds, ok := src.(int64)
	if !ok {
		return fmt.Errorf("sqltype.Timestamp.Scan: unsupported column type %T: %w", src, domain.ErrCorruptRecord)
	}

	*t = Timestamp(time.Unix(0, nanoseconds).UTC())

	return nil
}
