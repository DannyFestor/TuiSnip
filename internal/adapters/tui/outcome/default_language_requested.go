package outcome

import "github.com/DannyFestor/TuiSnip/internal/app/folder"

type DefaultLanguageRequested struct {
	Input folder.SetDefaultLanguageInput
}

func (DefaultLanguageRequested) isOutcome() {}
