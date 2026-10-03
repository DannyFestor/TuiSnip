package editoverlay

import "github.com/DannyFestor/TuiSnip/internal/domain"

type SaveFinished struct {
	Snippet domain.Snippet
	Err     error
}
