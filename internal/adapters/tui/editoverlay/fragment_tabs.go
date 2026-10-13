package editoverlay

import (
	"slices"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/fragmentstrip"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
)

type fragmentTabs struct {
	drafts []fragmentDraft
	strip  fragmentstrip.Strip
}

func newFragmentTabs(styles look.Styles, drafts []fragmentDraft, shown int) fragmentTabs {
	return fragmentTabs{drafts: nil, strip: fragmentstrip.New(styles)}.with(slices.Clone(drafts), shown)
}

func (t fragmentTabs) shownDraft() fragmentDraft {
	return t.drafts[t.shown()]
}

func (t fragmentTabs) shown() int {
	return t.strip.Shown()
}

func (t fragmentTabs) all() []fragmentDraft {
	return slices.Clone(t.drafts)
}

func (t fragmentTabs) single() bool {
	return len(t.drafts) == 1
}

func (t fragmentTabs) visible() bool {
	return t.strip.Visible()
}

func (t fragmentTabs) view(width int) string {
	return t.strip.View(width)
}

func (t fragmentTabs) restyled(styles look.Styles) fragmentTabs {
	t.strip = t.strip.Restyled(styles)

	return t
}

func (t fragmentTabs) withShown(draft fragmentDraft) fragmentTabs {
	drafts := t.all()
	drafts[t.shown()] = draft

	return t.with(drafts, t.shown())
}

func (t fragmentTabs) next() fragmentTabs {
	t.strip = t.strip.Next()

	return t
}

func (t fragmentTabs) previous() fragmentTabs {
	t.strip = t.strip.Previous()

	return t
}

func (t fragmentTabs) clicked(x int) fragmentTabs {
	t.strip = t.strip.Clicked(x)

	return t
}

func (t fragmentTabs) showing(index int) fragmentTabs {
	t.strip = t.strip.Showing(index)

	return t
}

func (t fragmentTabs) added(draft fragmentDraft) fragmentTabs {
	after := t.shown() + 1

	return t.with(slices.Insert(t.all(), after, draft), after)
}

func (t fragmentTabs) removed() fragmentTabs {
	return t.with(slices.Delete(t.all(), t.shown(), t.shown()+1), t.shown())
}

func (t fragmentTabs) movedBy(offset int) fragmentTabs {
	source, target := t.shown(), t.shown()+offset
	if target < 0 || target >= len(t.drafts) {
		return t
	}

	drafts := t.all()
	drafts[source], drafts[target] = drafts[target], drafts[source]

	return t.with(drafts, target)
}

func (t fragmentTabs) with(drafts []fragmentDraft, shown int) fragmentTabs {
	tabs := make([]fragmentstrip.Tab, 0, len(drafts))
	for _, draft := range drafts {
		tabs = append(tabs, draft.tab())
	}

	t.drafts = drafts
	t.strip = t.strip.OfTabs(tabs).Showing(min(shown, len(drafts)-1))

	return t
}
