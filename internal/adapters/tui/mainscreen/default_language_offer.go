package mainscreen

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/languagepicker"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const defaultLanguageTitle = "Default Language of "

func defaultLanguageOffer(filed domain.Folder, curated []value.Language) languagepicker.Offer {
	return languagepicker.Offer{
		Title:   defaultLanguageTitle + filed.Name().String(),
		Curated: curated,
		Current: filed.DefaultLanguage(),
		Picked: func(picked value.Language) outcome.Outcome {
			return outcome.DefaultLanguageRequested{
				Input: folder.SetDefaultLanguageInput{FolderID: filed.ID(), Language: picked.String()},
			}
		},
	}
}
