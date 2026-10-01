package snippet

import (
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type CopyResult struct {
	Delivery domain.CopyDelivery
	Content  value.Content
}
