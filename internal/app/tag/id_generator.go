package tag

import "github.com/DannyFestor/TuiSnip/internal/domain"

type IDGenerator interface {
	NewTagID() domain.TagID
}
