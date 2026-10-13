package fragmentstrip_test

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/fragmentstrip"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/screencell"
)

const wide = 80

func TestStrip_Visible(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		specs []testkit.FragmentSpec
		want  bool
	}{
		{name: "hides for one untitled Fragment", specs: []testkit.FragmentSpec{{Language: "Go"}}, want: false},
		{name: "shows for one titled Fragment", specs: []testkit.FragmentSpec{{Title: "main.go"}}, want: true},
		{
			name:  "shows for two untitled Fragments",
			specs: []testkit.FragmentSpec{{Language: "Go"}, {Language: "YAML"}},
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, stripOf(t, tt.specs...).Visible())
		})
	}
}

func TestStrip_View(t *testing.T) {
	t.Parallel()

	t.Run("labels each Fragment by its title, or its Language without one", func(t *testing.T) {
		t.Parallel()

		strip := stripOf(t, testkit.FragmentSpec{Title: "main.go"}, testkit.FragmentSpec{Language: "YAML"})

		assert.Equal(t, " main.go │ YAML ", ansi.Strip(strip.View(wide)))
	})

	t.Run("cuts a strip wider than the box off with an ellipsis", func(t *testing.T) {
		t.Parallel()

		strip := stripOf(t, testkit.FragmentSpec{Title: "main.go"}, testkit.FragmentSpec{Title: "handler.go"})

		assert.Equal(t, " main.go │ ha…", ansi.Strip(strip.View(14)))
	})
}

func TestStrip_Shown(t *testing.T) {
	t.Parallel()

	three := []testkit.FragmentSpec{{Title: "a.go"}, {Title: "b.go"}, {Title: "c.go"}}

	tests := []struct {
		name string
		move func(fragmentstrip.Strip) fragmentstrip.Strip
		want int
	}{
		{name: "starts on the first Fragment", move: unmoved, want: 0},
		{name: "next shows the following Fragment", move: fragmentstrip.Strip.Next, want: 1},
		{
			name: "next on the last Fragment wraps to the first",
			move: func(s fragmentstrip.Strip) fragmentstrip.Strip { return s.Next().Next().Next() },
			want: 0,
		},
		{name: "previous on the first Fragment wraps to the last", move: fragmentstrip.Strip.Previous, want: 2},
		{
			name: "first goes back to the first Fragment",
			move: func(s fragmentstrip.Strip) fragmentstrip.Strip { return s.Next().First() },
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.move(stripOf(t, three...)).Shown())
		})
	}

	t.Run("keeps the shown Fragment when the Fragments arrive again", func(t *testing.T) {
		t.Parallel()

		strip := stripOf(t, three...).Next().Next()

		assert.Equal(t, 2, strip.Of(fragments(t, three...)).Shown())
	})

	t.Run("shows the last Fragment when the shown one is gone", func(t *testing.T) {
		t.Parallel()

		strip := stripOf(t, three...).Next().Next()

		assert.Equal(t, 1, strip.Of(fragments(t, three[:2]...)).Shown())
	})
}

func TestStrip_Clicked(t *testing.T) {
	t.Parallel()

	strip := stripOf(t, testkit.FragmentSpec{Title: "main.go"}, testkit.FragmentSpec{Title: "go.mod"})
	view := strip.View(wide)

	t.Run("shows the clicked Fragment", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, 1, strip.Clicked(screencell.Find(t, view, "go.mod").X).Shown())
	})

	t.Run("ignores a click on a separator", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, 0, strip.Next().First().Clicked(screencell.Find(t, view, "│").X).Shown())
	})

	t.Run("ignores a click past the last label", func(t *testing.T) {
		t.Parallel()

		lastLabel := screencell.Find(t, view, "go.mod")

		assert.Equal(t, 1, strip.Next().Clicked(lastLabel.X+len("go.mod ")+1).Shown())
	})
}

func stripOf(t *testing.T, specs ...testkit.FragmentSpec) fragmentstrip.Strip {
	t.Helper()

	return fragmentstrip.New(look.NewStyles(look.SchemeDark)).Of(fragments(t, specs...))
}

func fragments(t *testing.T, specs ...testkit.FragmentSpec) []domain.Fragment {
	t.Helper()

	built := make([]domain.Fragment, 0, len(specs))
	for _, spec := range specs {
		built = append(built, testkit.Fragment(t, spec))
	}

	return built
}

func unmoved(strip fragmentstrip.Strip) fragmentstrip.Strip {
	return strip
}
