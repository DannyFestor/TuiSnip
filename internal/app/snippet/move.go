package snippet

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
		return nil, fmt.Errorf("snippet.NewMove: %w", err)
	}

	return &Move{repo: repo}, nil
}

func (m *Move) Run(ctx context.Context, input MoveInput) (domain.Snippet, error) {
	stored, err := m.repo.Find(ctx, input.SnippetID)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("snippet.Move: %w", err)
	}

	moved := stored.MoveTo(input.FolderID)

	err = m.repo.Move(ctx, moved)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("snippet.Move: %w", err)
	}

	return moved, nil
}
