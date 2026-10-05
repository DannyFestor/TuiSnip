package snippet

import (
	"errors"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type fields struct {
	title       value.Title
	description value.Description
	language    value.Language
	content     value.Content
}

func parseFields(title, description, language, content string) (fields, error) {
	parsedTitle, titleErr := value.NewTitle(title)
	parsedDescription, descriptionErr := value.NewDescription(description)
	parsedLanguage, languageErr := value.NewLanguage(language)
	parsedContent, contentErr := value.NewContent(content)

	parsed := fields{
		title:       parsedTitle,
		description: parsedDescription,
		language:    parsedLanguage,
		content:     parsedContent,
	}

	return parsed, errors.Join(
		domain.OnField(domain.FieldTitle, titleErr),
		domain.OnField(domain.FieldDescription, descriptionErr),
		domain.OnField(domain.FieldLanguage, languageErr),
		domain.OnField(domain.FieldContent, contentErr),
	)
}
