package config

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
)

const keyBindings = "bindings"

func layerBindings(defaults, user rawBindings) (map[Scope]map[Binding][]string, error) {
	layered := cloneBindings(defaults)
	overlayErr := overlayBindings(layered, user)
	typed, typedErr := typedBindings(layered)

	return typed, errors.Join(overlayErr, typedErr)
}

func cloneBindings(bindings rawBindings) rawBindings {
	clone := make(rawBindings, len(bindings))
	for scope, names := range bindings {
		clone[scope] = maps.Clone(names)
	}

	return clone
}

func overlayBindings(layered, user rawBindings) error {
	problems := make([]error, 0)

	for _, scope := range slices.Sorted(maps.Keys(user)) {
		known, ok := layered[scope]
		if !ok {
			problems = append(problems, atKey(scopeKey(scope), unknownScopeError()))

			continue
		}

		problems = append(problems, overlayScope(scope, known, user[scope]))
	}

	return errors.Join(problems...)
}

func overlayScope(scope string, known, user map[string][]string) error {
	problems := make([]error, 0)

	for _, name := range slices.Sorted(maps.Keys(user)) {
		if _, ok := known[name]; !ok {
			problems = append(problems, atKey(bindingKey(scope, name), unknownBindingError(scope, known)))

			continue
		}

		known[name] = user[name]
	}

	return errors.Join(problems...)
}

func typedBindings(raw rawBindings) (map[Scope]map[Binding][]string, error) {
	typed := make(map[Scope]map[Binding][]string, len(raw))
	problems := make([]error, 0)

	for scopeName, names := range raw {
		scope, err := ParseScope(scopeName)
		if err != nil {
			problems = append(problems, atKey(scopeKey(scopeName), err))

			continue
		}

		typed[scope] = make(map[Binding][]string, len(names))
		for name, keys := range names {
			binding, err := ParseBinding(name)
			problems = append(problems, atKey(bindingKey(scopeName, name), err))
			typed[scope][binding] = keys
		}
	}

	return typed, errors.Join(problems...)
}

func unknownScopeError() error {
	return fmt.Errorf("%w; use one of %s", errUnknownScope, strings.Join(ScopeNames(), ", "))
}

func unknownBindingError(scope string, known map[string][]string) error {
	return fmt.Errorf(
		"%w %s; use one of %s",
		errUnknownBinding,
		scope,
		strings.Join(slices.Sorted(maps.Keys(known)), ", "),
	)
}

func scopeKey(scope string) string {
	return keyBindings + "." + scope
}

func bindingKey(scope, name string) string {
	return scopeKey(scope) + "." + name
}
