package snippet

import (
	"errors"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type fields struct {
	title       value.Title
	description value.Description
	content     value.Content
}

func parseFields(title, description, content string) (fields, error) {
	parsedTitle, titleErr := value.NewTitle(title)
	parsedDescription, descriptionErr := value.NewDescription(description)
	parsedContent, contentErr := value.NewContent(content)

	return fields{title: parsedTitle, description: parsedDescription, content: parsedContent}, errors.Join(
		domain.OnField(domain.FieldTitle, titleErr),
		domain.OnField(domain.FieldDescription, descriptionErr),
		domain.OnField(domain.FieldContent, contentErr),
	)
}
