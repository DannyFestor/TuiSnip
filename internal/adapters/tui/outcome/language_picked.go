package outcome

import "github.com/DannyFestor/TuiSnip/internal/domain/value"

type LanguagePicked struct {
	Language value.Language
}

func (LanguagePicked) isOutcome() {}
