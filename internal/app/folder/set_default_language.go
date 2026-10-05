package folder

import (
	"context"
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type SetDefaultLanguage struct {
	repo  EditRepository
	clock Clock
}

func NewSetDefaultLanguage(repo EditRepository, clock Clock) (*SetDefaultLanguage, error) {
	err := errors.Join(
		domain.RequireDependency("repo", repo),
		domain.RequireDependency("clock", clock),
	)
	if err != nil {
		return nil, fmt.Errorf("folder.NewSetDefaultLanguage: %w", err)
	}

	return &SetDefaultLanguage{repo: repo, clock: clock}, nil
}

func (s *SetDefaultLanguage) Run(ctx context.Context, input SetDefaultLanguageInput) (domain.Folder, error) {
	language, err := value.NewLanguage(input.Language)
	if err != nil {
		return domain.Folder{}, fmt.Errorf("folder.SetDefaultLanguage: %w", domain.OnField(domain.FieldLanguage, err))
	}

	stored, err := s.repo.Find(ctx, input.FolderID)
	if err != nil {
		return domain.Folder{}, fmt.Errorf("folder.SetDefaultLanguage: %w", err)
	}

	changed, err := stored.WithDefaultLanguage(language, s.clock.Now())
	if err != nil {
		return domain.Folder{}, fmt.Errorf("folder.SetDefaultLanguage: %w", err)
	}

	err = s.repo.Update(ctx, changed)
	if err != nil {
		return domain.Folder{}, fmt.Errorf("folder.SetDefaultLanguage: %w", err)
	}

	return changed, nil
}
