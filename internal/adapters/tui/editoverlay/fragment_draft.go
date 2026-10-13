package editoverlay

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/fragmentstrip"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type fragmentDraft struct {
	id       domain.FragmentID
	title    string
	language value.Language
	content  string
}

func newFragmentDraft(language value.Language, content string) fragmentDraft {
	return fragmentDraft{id: domain.FragmentID{}, title: "", language: language, content: content}
}

func draftOf(fragment domain.Fragment) fragmentDraft {
	return fragmentDraft{
		id:       fragment.ID(),
		title:    fragment.Title().String(),
		language: fragment.Language(),
		content:  fragment.Content().String(),
	}
}

func draftsOf(fragments []domain.Fragment) []fragmentDraft {
	drafts := make([]fragmentDraft, 0, len(fragments))
	for _, fragment := range fragments {
		drafts = append(drafts, draftOf(fragment))
	}

	return drafts
}

func (d fragmentDraft) tab() fragmentstrip.Tab {
	title, err := value.NewFragmentTitle(d.title)
	if err != nil {
		return fragmentstrip.Tab{Title: value.FragmentTitle{}, Language: d.language}
	}

	return fragmentstrip.Tab{Title: title, Language: d.language}
}

func (d fragmentDraft) input() snippet.FragmentInput {
	return snippet.FragmentInput{ID: d.id, Title: d.title, Language: d.language.String(), Content: d.content}
}
