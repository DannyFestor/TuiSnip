package snippetslab

import (
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type folderRecord struct {
	UUID            string         `json:"uuid"`
	Title           string         `json:"title"`
	DefaultLanguage string         `json:"defaultLanguage"`
	Children        []folderRecord `json:"children"`
}

type folderParent struct {
	id       domain.FolderID
	language value.Language
}

func rootParent() folderParent {
	return folderParent{id: domain.FolderID{}, language: value.PlainText()}
}

func (c *conversion) folders(records []folderRecord, parent folderParent) ([]domain.Folder, error) {
	folders := make([]domain.Folder, 0, len(records))
	errs := make([]error, 0, len(records))

	for _, record := range records {
		subtree, err := c.subtree(record, parent)
		folders = append(folders, subtree...)
		errs = append(errs, err)
	}

	return folders, errors.Join(errs...)
}

func (c *conversion) subtree(record folderRecord, parent folderParent) ([]domain.Folder, error) {
	folder, err := c.folder(record, parent)
	if err != nil {
		_, childrenErr := c.folders(record.Children, parent)

		return nil, errors.Join(record.named(err), childrenErr)
	}

	children, err := c.folders(record.Children, folderParent{id: folder.ID(), language: folder.DefaultLanguage()})

	return append([]domain.Folder{folder}, children...), err
}

func (c *conversion) folder(record folderRecord, parent folderParent) (domain.Folder, error) {
	name, err := value.NewFolderName(record.Title)
	if err != nil {
		return domain.Folder{}, fmt.Errorf("folder name: %w", err)
	}

	language, known := record.defaultLanguage(parent.language)
	if !known {
		c.losses.FolderFallbacks = append(
			c.losses.FolderFallbacks,
			domain.FolderLanguageFallback{Folder: name, Lexer: record.DefaultLanguage},
		)
	}

	id := c.ids.NewFolderID()
	c.folderIDs[record.UUID] = id

	folder, err := domain.NewFolder(id, name, parent.id, language, c.now, c.now)
	if err != nil {
		return domain.Folder{}, fmt.Errorf("build folder: %w", err)
	}

	return folder, nil
}

func (r folderRecord) named(err error) error {
	return fmt.Errorf("folder %s: %w", r.UUID, err)
}

func (r folderRecord) defaultLanguage(inherited value.Language) (value.Language, bool) {
	if r.DefaultLanguage == "" {
		return inherited, true
	}

	return languageOfLexer(r.DefaultLanguage)
}
