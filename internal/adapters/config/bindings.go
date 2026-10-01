package config

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
)

const (
	keyBindings        = "bindings"
	problemsPerBinding = 2
)

func layerBindings(defaults, user rawBindings) (typedBindingMap, error) {
	layered := cloneBindings(defaults)
	overlayErr := overlayBindings(layered, user)
	typed, typedErr := typedBindings(layered)

	return typed, errors.Join(overlayErr, typedErr, checkBindings(typed))
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

func typedBindings(raw rawBindings) (typedBindingMap, error) {
	typed := make(typedBindingMap, len(raw))
	problems := make([]error, 0)

	for _, scopeName := range slices.Sorted(maps.Keys(raw)) {
		scope, err := ParseScope(scopeName)
		if err != nil {
			problems = append(problems, atKey(scopeKey(scopeName), err))

			continue
		}

		typed[scope], err = typedScope(scopeName, raw[scopeName])
		problems = append(problems, err)
	}

	return typed, errors.Join(problems...)
}

func typedScope(scopeName string, raw map[string][]string) (map[Binding][]Key, error) {
	typed := make(map[Binding][]Key, len(raw))
	problems := make([]error, 0, len(raw)*problemsPerBinding)

	for _, name := range slices.Sorted(maps.Keys(raw)) {
		path := bindingKey(scopeName, name)
		binding, bindingErr := ParseBinding(name)
		keys, keysErr := parseKeys(path, raw[name])
		problems = append(problems, atKey(path, bindingErr), keysErr)
		typed[binding] = keys
	}

	return typed, errors.Join(problems...)
}

func parseKeys(path string, written []string) ([]Key, error) {
	keys := make([]Key, 0, len(written))
	problems := make([]error, 0)

	for _, raw := range written {
		key, err := ParseKey(raw)
		if err != nil {
			problems = append(problems, atKey(path, err))

			continue
		}

		keys = append(keys, key)
	}

	return keys, errors.Join(problems...)
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
