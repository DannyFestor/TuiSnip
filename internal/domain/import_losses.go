package domain

import "github.com/DannyFestor/TuiSnip/internal/domain/value"

type ImportLosses struct {
	TrashedSnippets   int
	LostPins          []value.Title
	FolderFallbacks   []FolderLanguageFallback
	FragmentFallbacks []FragmentLanguageFallback
}
