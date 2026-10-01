package sqlite

import "fmt"

type NewerSchemaError struct {
	Database int64
	Known    int64
}

func (e NewerSchemaError) Error() string {
	return fmt.Sprintf(
		"The database was created by a newer TuiSnip (schema %d; this build knows %d). Upgrade TuiSnip.",
		e.Database, e.Known,
	)
}
