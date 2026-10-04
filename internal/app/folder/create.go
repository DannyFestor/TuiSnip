package folder

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type folderConstructor func(id domain.FolderID, name value.FolderName, now time.Time) (domain.Folder, error)

type Create struct {
	repo  CreateRepository
	ids   IDGenerator
	clock Clock
}

func NewCreate(repo CreateRepository, ids IDGenerator, clock Clock) (*Create, error) {
	err := errors.Join(
		domain.RequireDependency("repo", repo),
		domain.RequireDependency("ids", ids),
		domain.RequireDependency("clock", clock),
	)
	if err != nil {
		return nil, fmt.Errorf("folder.NewCreate: %w", err)
	}

	return &Create{repo: repo, ids: ids, clock: clock}, nil
}

func (c *Create) Run(ctx context.Context, input CreateInput) (domain.Folder, error) {
	name, err := parseName(input.Name)
	if err != nil {
		return domain.Folder{}, fmt.Errorf("folder.Create: %w", err)
	}

	folder, err := c.placed(ctx, input.ParentID, name)
	if err != nil {
		return domain.Folder{}, fmt.Errorf("folder.Create: %w", err)
	}

	err = c.repo.Insert(ctx, folder)
	if err != nil {
		return domain.Folder{}, fmt.Errorf("folder.Create: %w", err)
	}

	return folder, nil
}

func (c *Create) placed(ctx context.Context, parentID domain.FolderID, name value.FolderName) (domain.Folder, error) {
	build, err := c.constructorUnder(ctx, parentID)
	if err != nil {
		return domain.Folder{}, err
	}

	folder, err := build(c.ids.NewFolderID(), name, c.clock.Now())
	if err != nil {
		return domain.Folder{}, fmt.Errorf("new folder: %w", err)
	}

	return folder, nil
}

func (c *Create) constructorUnder(ctx context.Context, parentID domain.FolderID) (folderConstructor, error) {
	if parentID.IsNil() {
		return domain.NewFolderAtRoot, nil
	}

	parent, err := c.repo.Find(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("find parent: %w", err)
	}

	return func(id domain.FolderID, name value.FolderName, now time.Time) (domain.Folder, error) {
		return domain.NewFolderIn(parent, id, name, now)
	}, nil
}
