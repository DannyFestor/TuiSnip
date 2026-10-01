package value

import "errors"

const maxFolderNameRunes = 200

var (
	ErrBlankFolderName   = errors.New("value: folder name is blank")
	ErrFolderNameTooLong = errors.New("value: folder name is longer than 200 characters")
)

type FolderName = trimmedName[folderNameKind]

func NewFolderName(raw string) (FolderName, error) {
	return parseTrimmedName[folderNameKind](
		raw, nameRules{maxRunes: maxFolderNameRunes, errBlank: ErrBlankFolderName, errTooLong: ErrFolderNameTooLong},
	)
}
