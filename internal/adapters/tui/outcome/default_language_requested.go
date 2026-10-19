package outcome

import "github.com/DannyFestor/TuiSnip/internal/app/folder"

type DefaultLanguageRequested struct {
	Input folder.SetDefaultLanguageInput
	Name  string
}

func (DefaultLanguageRequested) isOutcome() {}
