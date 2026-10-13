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
	fragments   []FragmentInput
	newTags     []string
}

type fields struct {
	title       value.Title
	description value.Description
	fragments   []domain.FragmentEdit
	newTags     []value.TagName
}

func parseFields(raw rawFields) (fields, error) {
	parsedTitle, titleErr := value.NewTitle(raw.title)
	parsedDescription, descriptionErr := value.NewDescription(raw.description)
	parsedFragments, fragmentsErr := parseFragments(raw.fragments)
	parsedNewTags, newTagsErr := parseTagNames(raw.newTags)

	parsed := fields{
		title:       parsedTitle,
		description: parsedDescription,
		fragments:   parsedFragments,
		newTags:     parsedNewTags,
	}

	return parsed, errors.Join(
		domain.OnField(domain.FieldTitle, titleErr),
		domain.OnField(domain.FieldDescription, descriptionErr),
		fragmentsErr,
		newTagsErr,
	)
}

func parseFragments(raw []FragmentInput) ([]domain.FragmentEdit, error) {
	parsed := make([]domain.FragmentEdit, 0, len(raw))
	fragmentErrs := make([]error, 0, len(raw))

	for position, input := range raw {
		fragment, err := parseFragment(input)
		parsed = append(parsed, fragment)
		fragmentErrs = append(fragmentErrs, domain.InFragment(position, err))
	}

	return parsed, errors.Join(fragmentErrs...)
}

func parseFragment(raw FragmentInput) (domain.FragmentEdit, error) {
	title, titleErr := value.NewFragmentTitle(raw.Title)
	language, languageErr := value.NewLanguage(raw.Language)
	content, contentErr := value.NewContent(raw.Content)

	parsed := domain.FragmentEdit{ID: raw.ID, Title: title, Language: language, Content: content}

	return parsed, errors.Join(
		domain.OnField(domain.FieldFragmentTitle, titleErr),
		domain.OnField(domain.FieldLanguage, languageErr),
		domain.OnField(domain.FieldContent, contentErr),
	)
}

func (f fields) newFragments(ids IDGenerator, now time.Time) ([]domain.Fragment, error) {
	fragments := make([]domain.Fragment, 0, len(f.fragments))

	for _, parsed := range f.fragments {
		fragment, err := domain.NewFragment(
			ids.NewFragmentID(),
			parsed.Title,
			parsed.Language,
			parsed.Content,
			now,
			now,
		)
		if err != nil {
			return nil, fmt.Errorf("new fragment: %w", err)
		}

		fragments = append(fragments, fragment)
	}

	return fragments, nil
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
