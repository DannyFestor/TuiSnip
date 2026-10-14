package domain

import "github.com/DannyFestor/TuiSnip/internal/domain/value"

type FolderLanguageFallback struct {
	Folder value.FolderName
	Lexer  string
}
