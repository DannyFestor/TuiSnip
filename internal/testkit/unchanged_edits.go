package testkit

import "github.com/DannyFestor/TuiSnip/internal/domain"

func UnchangedEdits(snippet domain.Snippet) []domain.FragmentEdit {
	edits := make([]domain.FragmentEdit, 0, len(snippet.Fragments()))
	for _, fragment := range snippet.Fragments() {
		edits = append(edits, domain.FragmentEdit{
			ID:       fragment.ID(),
			Title:    fragment.Title(),
			Language: fragment.Language(),
			Content:  fragment.Content(),
		})
	}

	return edits
}
