package domain

import "github.com/DannyFestor/TuiSnip/internal/domain/value"

type ImportLosses struct {
	LostPins          []value.Title
	LostGistLinks     []value.Title
	LostAttachments   []value.Title
	FolderFallbacks   []FolderLanguageFallback
	FragmentFallbacks []FragmentLanguageFallback
}
