package editoverlay_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/editoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/screencolour"
)

const (
	firstGoFragment  = "func main() {}"
	secondGoFragment = "var x = 1"
)

func TestEditing_highlighting(t *testing.T) {
	t.Parallel()

	t.Run("shows the Snippet pane's colours, block comments included", func(t *testing.T) {
		t.Parallel()

		const content = "/* first\nsecond */\nfunc main() {}"

		screen := editingStored(t, storedSnippet(t, content))
		screen.RunCommands()

		for _, text := range []string{"second */", "func"} {
			assert.Equal(t,
				goInSnippetPane(look.SchemeDark, content).Foreground(t, text),
				screencolour.Foreground(t, screen.StyledScreen(), text),
				text,
			)
		}
	})

	t.Run("keeps the old colours until typing pauses, then updates them", func(t *testing.T) {
		t.Parallel()

		const content = "func main() {}\n"

		screen := editingStored(t, storedSnippet(t, content))
		screen.RunCommands()
		screen.Press(enterContent()...)
		screen.Press(keypress.Typed("// note")...)

		assert.Equal(t,
			goInSnippetPane(look.SchemeDark, content).Foreground(t, "func"),
			screencolour.Foreground(t, screen.StyledScreen(), "func"),
		)
		assert.Empty(t, screencolour.Foreground(t, screen.StyledScreen(), "// note"))

		screen.RunCommands()

		assert.Equal(t,
			goInSnippetPane(look.SchemeDark, content+"// note").Foreground(t, "// note"),
			screencolour.Foreground(t, screen.StyledScreen(), "// note"),
		)
	})

	t.Run("shows plain text without colours", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, storedSnippetIn(t, "plaintext", "func main() {}"))
		screen.RunCommands()

		assert.Empty(t, screencolour.Foreground(t, screen.StyledScreen(), "func"))
	})

	t.Run("colours the Content again in a picked Language", func(t *testing.T) {
		t.Parallel()

		const content = "func main() {}"

		screen := editingStored(t, storedSnippetIn(t, "plaintext", content))
		screen.RunCommands()
		screen.Offer(outcome.LanguagePicked{Language: language(t, "Go")})
		screen.RunCommands()

		assert.Equal(t,
			goInSnippetPane(look.SchemeDark, content).Foreground(t, "func"),
			screencolour.Foreground(t, screen.StyledScreen(), "func"),
		)
	})

	t.Run("shows no colours at once when plain text is picked", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, storedSnippet(t, "func main() {}"))
		screen.RunCommands()
		screen.Offer(outcome.LanguagePicked{Language: language(t, "plaintext")})

		assert.Empty(t, screencolour.Foreground(t, screen.StyledScreen(), "func"))
	})

	t.Run("shows another Fragment without the colours of the one before", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, twoGoFragments(t))
		screen.RunCommands()
		screen.Press(nextFragment())

		assert.Empty(t, screencolour.Foreground(t, screen.StyledScreen(), "var"))

		screen.RunCommands()

		assert.Equal(t,
			goInSnippetPane(look.SchemeDark, secondGoFragment).Foreground(t, "var"),
			screencolour.Foreground(t, screen.StyledScreen(), "var"),
		)
	})

	t.Run("drops the colours lexed for a Fragment shown before", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, twoGoFragments(t))
		screen.RunQueuedCommands()
		screen.Press(nextFragment())
		screen.RunQueuedCommands()

		assert.Empty(t, screencolour.Foreground(t, screen.StyledScreen(), "var"))
	})

	t.Run("colours the Fragment a refused save shows", func(t *testing.T) {
		t.Parallel()

		titleTooLong := domain.OnField(domain.FieldFragmentTitle, value.ErrFragmentTitleTooLong)
		screen := editingStored(t, twoGoFragments(t))
		screen.Press(save())
		screen.RunCommands()
		screen.Send(editoverlay.SaveFinished{
			Snippet: domain.Snippet{},
			Err:     fmt.Errorf("snippet.Update: %w", domain.InFragment(1, titleTooLong)),
		})
		screen.RunCommands()

		assert.Equal(t,
			goInSnippetPane(look.SchemeDark, secondGoFragment).Foreground(t, "var"),
			screencolour.Foreground(t, screen.StyledScreen(), "var"),
		)
	})

	t.Run("keeps the colours of a Fragment moved along the strip", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, twoGoFragments(t))
		screen.RunCommands()
		screen.Press(moveFragmentRight())

		assert.Equal(t,
			goInSnippetPane(look.SchemeDark, firstGoFragment).Foreground(t, "func"),
			screencolour.Foreground(t, screen.StyledScreen(), "func"),
		)
	})

	t.Run("colours the Content in the code style of a new scheme", func(t *testing.T) {
		t.Parallel()

		const content = "func main() {}"

		screen := editingStored(t, storedSnippet(t, content))
		screen.RunCommands()
		screen.Send(look.Restyled{Styles: look.NewStyles(look.SchemeLight)})
		screen.RunCommands()

		assert.Equal(t,
			goInSnippetPane(look.SchemeLight, content).Foreground(t, "func"),
			screencolour.Foreground(t, screen.StyledScreen(), "func"),
		)
	})
}

func twoGoFragments(t *testing.T) domain.Snippet {
	t.Helper()

	return storedSnippetOf(t,
		testkit.FragmentSpec{Language: "Go", Content: firstGoFragment},
		testkit.FragmentSpec{Language: "Go", Content: secondGoFragment},
	)
}

func goInSnippetPane(scheme look.Scheme, content string) screencolour.InSnippetPane {
	return screencolour.InSnippetPane{Content: content, Language: "Go", Scheme: scheme}
}
