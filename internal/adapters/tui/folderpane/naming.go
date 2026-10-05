package folderpane

import (
	"errors"
	"fmt"
	"slices"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	newFolderCount = "0"
	blankNameText  = "Folder name is blank"
	longNameText   = "Folder name is longer than 200 characters"
)

type naming interface {
	placed(at row, index int) placement
	requested(name string) []outcome.Outcome
}

type placement struct {
	index  int
	insert bool
	prefix string
	meta   string
}

type newFolder struct {
	parentID domain.FolderID
}

type renamedFolder struct {
	folderID domain.FolderID
}

func (newFolder) placed(at row, index int) placement {
	return placement{index: index + 1, insert: true, prefix: at.childIndent + leafMarker, meta: newFolderCount}
}

func (n newFolder) requested(name string) []outcome.Outcome {
	return []outcome.Outcome{
		outcome.FolderCreateRequested{Input: folder.CreateInput{Name: name, ParentID: n.parentID}},
	}
}

func (renamedFolder) placed(at row, index int) placement {
	return placement{index: index, insert: false, prefix: at.prefix, meta: at.count()}
}

func (r renamedFolder) requested(name string) []outcome.Outcome {
	return []outcome.Outcome{
		outcome.FolderRenameRequested{Input: folder.RenameInput{FolderID: r.folderID, Name: name}},
	}
}

func (p placement) into(lines []line, field string) []line {
	shown := line{text: p.prefix + field, meta: p.meta}
	if p.insert {
		return slices.Insert(lines, p.index, shown)
	}

	lines[p.index] = shown

	return lines
}

func validFolderName(raw string) error {
	_, err := value.NewFolderName(raw)
	if err != nil {
		return fmt.Errorf("folderpane: %w", err)
	}

	return nil
}

func refusalText(err error) string {
	switch {
	case errors.Is(err, value.ErrBlankFolderName):
		return blankNameText
	case errors.Is(err, value.ErrFolderNameTooLong):
		return longNameText
	}

	return look.FailureText
}
