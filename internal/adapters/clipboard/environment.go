package clipboard

import (
	"slices"
	"strings"
)

type environment []string

func (e environment) get(key string) string {
	value := ""
	prefix := key + "="

	for _, entry := range e {
		if after, found := strings.CutPrefix(entry, prefix); found {
			value = after
		}
	}

	return value
}

func (e environment) with(key, value string) environment {
	return append(slices.Clip(e), key+"="+value)
}
