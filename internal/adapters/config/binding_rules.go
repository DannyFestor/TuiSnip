package config

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

type typedBindingMap = map[Scope]map[Binding][]Key

type keyOwner struct {
	scope   Scope
	binding Binding
}

type boundKey struct {
	owner keyOwner
	key   Key
}

type scopedKey struct {
	scope Scope
	key   Key
}

func checkBindings(bindings typedBindingMap) error {
	bound := boundKeys(bindings)

	return errors.Join(
		checkForcedQuit(bound),
		checkTypesText(bound),
		checkSameScope(bound),
		checkGlobalAgainstPanes(bound),
	)
}

func boundKeys(bindings typedBindingMap) []boundKey {
	bound := make([]boundKey, 0)

	for _, scope := range slices.Sorted(maps.Keys(bindings)) {
		for _, binding := range slices.Sorted(maps.Keys(bindings[scope])) {
			for _, key := range bindings[scope][binding] {
				bound = append(bound, boundKey{owner: keyOwner{scope: scope, binding: binding}, key: key})
			}
		}
	}

	return bound
}

func checkForcedQuit(bound []boundKey) error {
	problems := make([]error, 0)

	for _, candidate := range bound {
		if candidate.key == ForcedQuitKey() {
			problems = append(problems, candidate.problem(errForcedQuit))
		}
	}

	return errors.Join(problems...)
}

func checkTypesText(bound []boundKey) error {
	problems := make([]error, 0)

	for _, candidate := range bound {
		if candidate.owner.scope.isTextEntry() && candidate.key.typesText() {
			problems = append(
				problems,
				candidate.problem(fmt.Errorf("%w %s; add a modifier", errTypesText, candidate.owner.scope)),
			)
		}
	}

	return errors.Join(problems...)
}

func checkSameScope(bound []boundKey) error {
	firstOwners := make(map[scopedKey]keyOwner, len(bound))
	problems := make([]error, 0)

	for _, candidate := range bound {
		scoped := scopedKey{scope: candidate.owner.scope, key: candidate.key}

		first, taken := firstOwners[scoped]
		if !taken {
			firstOwners[scoped] = candidate.owner

			continue
		}

		if first != candidate.owner {
			problems = append(problems, candidate.alsoBoundTo(first))
		}
	}

	return errors.Join(problems...)
}

func checkGlobalAgainstPanes(bound []boundKey) error {
	paneOwners := make(map[Key]keyOwner, len(bound))
	problems := make([]error, 0)

	for _, candidate := range bound {
		if _, taken := paneOwners[candidate.key]; !taken && candidate.owner.scope.isPane() {
			paneOwners[candidate.key] = candidate.owner
		}
	}

	for _, candidate := range bound {
		owner, taken := paneOwners[candidate.key]
		if taken && candidate.owner.scope == ScopeGlobal {
			problems = append(problems, candidate.alsoBoundTo(owner))
		}
	}

	return errors.Join(problems...)
}

func (b boundKey) problem(err error) error {
	return atKey(b.owner.path(), fmt.Errorf("%q %w", b.key.String(), err))
}

func (b boundKey) alsoBoundTo(other keyOwner) error {
	return b.problem(fmt.Errorf("%w %s", errAlsoBound, other.path()))
}

func (o keyOwner) path() string {
	return bindingKey(o.scope.String(), o.binding.String())
}
