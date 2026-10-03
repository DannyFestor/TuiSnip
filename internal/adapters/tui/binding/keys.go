package binding

import "charm.land/bubbles/v2/key"

type Keys map[Scope]map[string][]string

func (k Keys) For(scope Scope) Set {
	bound := make(map[string]key.Binding)

	for _, row := range Rows() {
		if row.Scope == scope {
			bound[row.Name] = k.resolved(rowRef{scope: row.Scope, name: row.Name}, row.Label)
		}
	}

	return Set{bound: bound, hints: k.resolvedHints(statusHintList(scope))}
}

func (k Keys) EmptyListHints() []key.Binding {
	return k.resolvedHints(emptyListHintList())
}

func (k Keys) resolvedHints(hints []labelledRef) []key.Binding {
	resolved := make([]key.Binding, 0, len(hints))
	for _, hint := range hints {
		resolved = append(resolved, k.resolved(hint.row, hint.label))
	}

	return resolved
}

func (k Keys) resolved(ref rowRef, label string) key.Binding {
	keys := k[ref.scope][ref.name]
	resolved := key.NewBinding(key.WithKeys(keys...))

	if len(keys) == 0 {
		resolved.SetEnabled(false)

		return resolved
	}

	resolved.SetHelp(keys[0], label)

	return resolved
}
