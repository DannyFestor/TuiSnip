//go:build feature

package feature_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/app/trash"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const plainText = "plaintext"

func create(t *testing.T, app *bootstrap.App, in snippet.CreateInput) domain.Snippet {
	t.Helper()

	created, err := app.Create.Run(t.Context(), in)
	require.NoError(t, err)

	return created
}

func trashItems(t *testing.T, app *bootstrap.App) []trash.Item {
	t.Helper()

	listed, err := app.ListTrash.Run(t.Context(), trash.ListInput{})
	require.NoError(t, err)

	return listed
}

func plainFragment(content string) []snippet.FragmentInput {
	return fragmentIn(plainText, content)
}

func fragmentIn(language, content string) []snippet.FragmentInput {
	return []snippet.FragmentInput{{ID: domain.FragmentID{}, Title: "", Language: language, Content: content}}
}

func fragmentInputsOf(stored domain.Snippet) []snippet.FragmentInput {
	inputs := make([]snippet.FragmentInput, 0, len(stored.Fragments()))
	for _, fragment := range stored.Fragments() {
		inputs = append(inputs, snippet.FragmentInput{
			ID:       fragment.ID(),
			Title:    fragment.Title().String(),
			Language: fragment.Language().String(),
			Content:  fragment.Content().String(),
		})
	}

	return inputs
}
