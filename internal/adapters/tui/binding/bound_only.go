package binding

import "slices"

func BoundOnly(firstKeys ...string) []string {
	return slices.DeleteFunc(firstKeys, func(firstKey string) bool { return firstKey == "" })
}
