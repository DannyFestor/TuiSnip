package browse

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type TagList struct {
	tags    TagLister
	counter TagSnippetCounter
}

func NewTagList(tags TagLister, counter TagSnippetCounter) (*TagList, error) {
	err := errors.Join(
		domain.RequireDependency("tags", tags),
		domain.RequireDependency("counter", counter),
	)
	if err != nil {
		return nil, fmt.Errorf("browse.NewTagList: %w", err)
	}

	return &TagList{tags: tags, counter: counter}, nil
}

func (l *TagList) Run(ctx context.Context, _ TagListInput) ([]TagCount, error) {
	tags, err := l.tags.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("browse.TagList: %w", err)
	}

	counts, err := l.counter.CountByTag(ctx)
	if err != nil {
		return nil, fmt.Errorf("browse.TagList: %w", err)
	}

	slices.SortFunc(tags, domain.CompareTags)

	counted := make([]TagCount, 0, len(tags))
	for _, tag := range tags {
		counted = append(counted, TagCount{Tag: tag, SnippetCount: counts[tag.ID()]})
	}

	return counted, nil
}
