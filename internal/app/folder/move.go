package folder

import (
	"context"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Move struct {
	repo MoveRepository
}

func NewMove(repo MoveRepository) (*Move, error) {
	err := domain.RequireDependency("repo", repo)
	if err != nil {
		return nil, fmt.Errorf("folder.NewMove: %w", err)
	}

	return &Move{repo: repo}, nil
}

func (m *Move) Run(ctx context.Context, input MoveInput) (domain.Folder, error) {
	stored, err := m.repo.Find(ctx, input.FolderID)
	if err != nil {
		return domain.Folder{}, fmt.Errorf("folder.Move: %w", err)
	}

	moved, err := m.moved(ctx, stored, input.ParentID)
	if err != nil {
		return domain.Folder{}, fmt.Errorf("folder.Move: %w", err)
	}

	err = m.repo.Move(ctx, moved)
	if err != nil {
		return domain.Folder{}, fmt.Errorf("folder.Move: %w", err)
	}

	return moved, nil
}

func (m *Move) moved(ctx context.Context, stored domain.Folder, parentID domain.FolderID) (domain.Folder, error) {
	if parentID.IsNil() {
		return stored.MoveToRoot(), nil
	}

	descendantIDs, err := m.repo.ListDescendantIDs(ctx, stored.ID())
	if err != nil {
		return domain.Folder{}, fmt.Errorf("list descendants: %w", err)
	}

	moved, err := stored.MoveUnder(parentID, descendantIDs)
	if err != nil {
		return domain.Folder{}, fmt.Errorf("move under: %w", err)
	}

	return moved, nil
}
