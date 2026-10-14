package library

import "github.com/DannyFestor/TuiSnip/internal/domain"

type SnippetsLabReport struct {
	Counts Counts
	Losses domain.ImportLosses
}
