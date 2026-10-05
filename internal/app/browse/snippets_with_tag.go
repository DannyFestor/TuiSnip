package browse

import (
	"context"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SnippetsWithTag struct {
	lister TagSnippetLister
}

func NewSnippetsWithTag(lister TagSnippetLister) (*SnippetsWithTag, error) {
	err := domain.RequireDependency("lister", lister)
	if err != nil {
		return nil, fmt.Errorf("browse.NewSnippetsWithTag: %w", err)
	}

	return &SnippetsWithTag{lister: lister}, nil
}

func (s *SnippetsWithTag) Run(ctx context.Context, in SnippetsWithTagInput) ([]domain.Snippet, error) {
	return listedBy("browse.SnippetsWithTag")(s.lister.ListWithTag(ctx, in.TagID, in.Order))
}
