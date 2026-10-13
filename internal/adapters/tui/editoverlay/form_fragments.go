package editoverlay

import (
	"errors"
	"slices"

	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

func (f form) fragments() fragmentTabs {
	return f.tabs.withShown(f.shownDraft())
}

func (f form) shownDraft() fragmentDraft {
	draft := f.tabs.shownDraft()
	draft.title = f.fragmentTitle.Value()
	draft.language = f.language
	draft.content = f.contentValue()

	return draft
}

func (f form) shownFragmentName() string {
	return f.shownDraft().tab().Label()
}

func (f form) contentValue() string {
	if f.readOnly.held() {
		return f.readOnly.content
	}

	return f.content.Value()
}

func (f form) fragmentKeyPressed(msg tea.KeyPressMsg) (form, request, bool) {
	fields := f.keys.fields

	switch {
	case fields.Matches(msg, binding.NextFragment):
		return f.withTabsChanged(fragmentTabs.next), requestNothing, true
	case fields.Matches(msg, binding.PrevFragment):
		return f.withTabsChanged(fragmentTabs.previous), requestNothing, true
	case fields.Matches(msg, binding.MoveFragmentLeft):
		return f.withFragmentMovedBy(-1), requestNothing, true
	case fields.Matches(msg, binding.MoveFragmentRight):
		return f.withFragmentMovedBy(1), requestNothing, true
	case fields.Matches(msg, binding.AddFragment):
		return f, requestAddFragment, true
	case fields.Matches(msg, binding.RemoveFragment):
		return f, f.removalRequest(), true
	}

	return f, requestNothing, false
}

func (f form) removalRequest() request {
	if f.tabs.single() {
		return requestRefuseRemovingLastFragment
	}

	return requestRemoveFragment
}

func (f form) withFragmentAdded(language value.Language) form {
	return f.withTabsChanged(func(tabs fragmentTabs) fragmentTabs {
		return tabs.added(newFragmentDraft(language, ""))
	})
}

func (f form) withShownFragmentRemoved() form {
	return f.withTabsChanged(fragmentTabs.removed)
}

func (f form) withFragmentMovedBy(offset int) form {
	return f.withTabsChanged(func(tabs fragmentTabs) fragmentTabs { return tabs.movedBy(offset) })
}

func (f form) withFragmentClicked(x int) form {
	return f.withTabsChanged(func(tabs fragmentTabs) fragmentTabs { return tabs.clicked(x) })
}

func (f form) showingFragmentOf(err error) form {
	fragmentErr, ok := errors.AsType[domain.FragmentError](err)
	if !ok {
		return f
	}

	return f.withTabsChanged(func(tabs fragmentTabs) fragmentTabs { return tabs.showing(fragmentErr.Position) })
}

func (f form) withTabsChanged(change func(fragmentTabs) fragmentTabs) form {
	next := f
	next.tabs = change(f.fragments())
	next.invalid = slices.DeleteFunc(slices.Clone(f.invalid), isFragmentField)

	return next.withDraft(next.tabs.shownDraft()).resized(next.outer)
}

func (f form) withDraft(draft fragmentDraft) form {
	next := f
	next.fragmentTitle.SetValue(draft.title)
	next.language = draft.language
	next.readOnly = readOnlyIfUneditable(draft.content, draft.language, f.codeStyle)
	next.content.SetValue("")

	return next.withContent(draft.content)
}

func (f form) stripRows() int {
	if f.fragments().visible() {
		return 1
	}

	return 0
}
