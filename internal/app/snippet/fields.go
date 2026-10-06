package snippet

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type rawFields struct {
	title       string
	description string
	language    string
	content     string
	newTags     []string
}

type fields struct {
	title       value.Title
	description value.Description
	language    value.Language
	content     value.Content
	newTags     []value.TagName
}

func parseFields(raw rawFields) (fields, error) {
	parsedTitle, titleErr := value.NewTitle(raw.title)
	parsedDescription, descriptionErr := value.NewDescription(raw.description)
	parsedLanguage, languageErr := value.NewLanguage(raw.language)
	parsedContent, contentErr := value.NewContent(raw.content)
	parsedNewTags, newTagsErr := parseTagNames(raw.newTags)

	parsed := fields{
		title:       parsedTitle,
		description: parsedDescription,
		language:    parsedLanguage,
		content:     parsedContent,
		newTags:     parsedNewTags,
	}

	return parsed, errors.Join(
		domain.OnField(domain.FieldTitle, titleErr),
		domain.OnField(domain.FieldDescription, descriptionErr),
		domain.OnField(domain.FieldLanguage, languageErr),
		domain.OnField(domain.FieldContent, contentErr),
		newTagsErr,
	)
}

func parseTagNames(raw []string) ([]value.TagName, error) {
	names := make([]value.TagName, 0, len(raw))
	nameErrs := make([]error, 0)

	for _, rawName := range raw {
		name, err := value.NewTagName(rawName)
		if err != nil {
			nameErrs = append(nameErrs, domain.OnField(domain.FieldTagName, err))

			continue
		}

		names = append(names, name)
	}

	return names, errors.Join(nameErrs...)
}

func (f fields) tagsBeside(stored []domain.Tag, ids IDGenerator, now time.Time) ([]domain.Tag, error) {
	tags := slices.Grow(slices.Clone(stored), len(f.newTags))

	for _, name := range f.newTags {
		created, err := domain.NewTag(ids.NewTagID(), name, now, now)
		if err != nil {
			return nil, fmt.Errorf("new tag: %w", err)
		}

		tags = append(tags, created)
	}

	return tags, nil
}
