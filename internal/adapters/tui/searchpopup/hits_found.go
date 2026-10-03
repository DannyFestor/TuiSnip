package searchpopup

import "github.com/DannyFestor/TuiSnip/internal/domain"

type HitsFound struct {
	Text string
	Hits []domain.SearchHit
}
