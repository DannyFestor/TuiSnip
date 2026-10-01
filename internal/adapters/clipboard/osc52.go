package clipboard

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type OSC52 struct{}

func NewOSC52() *OSC52 {
	return &OSC52{}
}

func (*OSC52) Copy(context.Context, string) (domain.CopyDelivery, error) {
	return domain.CopyDeliverySentToTerminal, nil
}
