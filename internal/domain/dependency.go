package domain

import "fmt"

func RequireDependency(name string, dependency any) error {
	if dependency == nil {
		return fmt.Errorf("%s: %w", name, ErrMissingDependency)
	}

	return nil
}
