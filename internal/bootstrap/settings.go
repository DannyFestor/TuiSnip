package bootstrap

import (
	"time"

	"github.com/DannyFestor/TuiSnip/internal/adapters/config"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
)

func SettingsFrom(cfg config.Config, location *time.Location) tui.Settings {
	keys := keyStrings(cfg.Bindings)

	return tui.Settings{
		Global:      globalKeyMap(keys.in(config.ScopeGlobal)),
		Folders:     tui.FoldersKeyMap{NewFolder: keys.in(config.ScopeFolders)(config.BindingNewFolder)},
		SnippetList: tui.SnippetListKeyMap{Copy: keys.in(config.ScopeSnippetList)(config.BindingCopy)},
		SnippetPane: tui.SnippetPaneKeyMap{Copy: keys.in(config.ScopeSnippetPane)(config.BindingCopy)},
		Editor:      editorKeyMap(keys.in(config.ScopeEditor)),
		Content:     contentKeyMap(keys.in(config.ScopeContent)),
		Search:      searchKeyMap(keys.in(config.ScopeSearch)),
		Confirm:     confirmKeyMap(keys.in(config.ScopeConfirm)),
		Location:    location,
	}
}

type keyStrings map[config.Scope]map[config.Binding][]config.Key

type scopeKeys func(binding config.Binding) []string

func (k keyStrings) in(scope config.Scope) scopeKeys {
	return func(binding config.Binding) []string {
		bound := k[scope][binding]
		written := make([]string, 0, len(bound))

		for _, key := range bound {
			written = append(written, key.String())
		}

		return written
	}
}

func globalKeyMap(keys scopeKeys) tui.GlobalKeyMap {
	return tui.GlobalKeyMap{
		Quit:         keys(config.BindingQuit),
		Help:         keys(config.BindingHelp),
		Search:       keys(config.BindingSearch),
		NewSnippet:   keys(config.BindingNewSnippet),
		Capture:      keys(config.BindingCapture),
		FocusNext:    keys(config.BindingFocusNext),
		FocusPrev:    keys(config.BindingFocusPrev),
		FocusRight:   keys(config.BindingFocusRight),
		FocusLeft:    keys(config.BindingFocusLeft),
		FocusFolders: keys(config.BindingFocusFolders),
		FocusTags:    keys(config.BindingFocusTags),
		FocusList:    keys(config.BindingFocusList),
		FocusSnippet: keys(config.BindingFocusSnippet),
		Open:         keys(config.BindingOpen),
		Back:         keys(config.BindingBack),
		Down:         keys(config.BindingDown),
		Up:           keys(config.BindingUp),
		Top:          keys(config.BindingTop),
		Bottom:       keys(config.BindingBottom),
		PageDown:     keys(config.BindingPageDown),
		PageUp:       keys(config.BindingPageUp),
	}
}

func editorKeyMap(keys scopeKeys) tui.EditorKeyMap {
	return tui.EditorKeyMap{
		Save:      keys(config.BindingSave),
		Cancel:    keys(config.BindingCancel),
		NextField: keys(config.BindingNextField),
		PrevField: keys(config.BindingPrevField),
		OpenField: keys(config.BindingOpenField),
	}
}

func contentKeyMap(keys scopeKeys) tui.ContentKeyMap {
	return tui.ContentKeyMap{Save: keys(config.BindingSave), Leave: keys(config.BindingLeave)}
}

func searchKeyMap(keys scopeKeys) tui.SearchKeyMap {
	return tui.SearchKeyMap{
		Down:   keys(config.BindingDown),
		Up:     keys(config.BindingUp),
		Accept: keys(config.BindingAccept),
		Copy:   keys(config.BindingCopy),
		Cancel: keys(config.BindingCancel),
	}
}

func confirmKeyMap(keys scopeKeys) tui.ConfirmKeyMap {
	return tui.ConfirmKeyMap{Yes: keys(config.BindingYes), No: keys(config.BindingNo)}
}
