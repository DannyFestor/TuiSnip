package browse

import (
	"context"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SnippetsInFolder struct {
	lister SnippetLister
}

func NewSnippetsInFolder(lister SnippetLister) (*SnippetsInFolder, error) {
	err := domain.RequireDependency("lister", lister)
	if err != nil {
		return nil, fmt.Errorf("browse.NewSnippetsInFolder: %w", err)
	}

	return &SnippetsInFolder{lister: lister}, nil
}

func (s *SnippetsInFolder) Run(ctx context.Context, in SnippetsInFolderInput) ([]domain.Snippet, error) {
	snippets, err := s.lister.ListInFolder(ctx, in.FolderID)
	if err != nil {
		return nil, fmt.Errorf("browse.SnippetsInFolder: %w", err)
	}

	return snippets, nil
}
