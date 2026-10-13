package editoverlay

import "github.com/DannyFestor/TuiSnip/internal/domain"

type ExternallyEdited struct {
	Browsed  BrowsedSnippet
	Fragment domain.Fragment
	Content  string
}
