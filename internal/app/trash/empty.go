package trash

import (
	"context"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Empty struct {
	emptier Emptier
}

func NewEmpty(emptier Emptier) (*Empty, error) {
	err := domain.RequireDependency("emptier", emptier)
	if err != nil {
		return nil, fmt.Errorf("trash.NewEmpty: %w", err)
	}

	return &Empty{emptier: emptier}, nil
}

func (e *Empty) Run(ctx context.Context, input EmptyInput) (domain.TrashEmptied, error) {
	emptied, err := e.emptier.Empty(ctx, idsByKind(input.Items))
	if err != nil {
		return domain.TrashEmptied{}, fmt.Errorf("trash.Empty: %w", err)
	}

	return emptied, nil
}

func idsByKind(items []ItemID) domain.TrashItemIDs {
	var sorted domain.TrashItemIDs

	for _, item := range items {
		if snippetID, ok := item.Snippet(); ok {
			sorted.Snippets = append(sorted.Snippets, snippetID)

			continue
		}

		folderID, _ := item.Folder()
		sorted.Folders = append(sorted.Folders, folderID)
	}

	return sorted
}
