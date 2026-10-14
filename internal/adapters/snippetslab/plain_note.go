package snippetslab

import (
	"cmp"
	"slices"
	"strings"
	"unicode/utf16"
)

// SnippetsLab counts note ranges in UTF-16 code units, so the note is cut as UTF-16
// and not as Go's bytes or runes.
func plainNote(note string, attributes []noteAttribute) (string, error) {
	units := utf16.Encode([]rune(note))

	var plain strings.Builder

	next := 0

	for _, link := range linksInOrder(attributes) {
		span, err := link.span(next, len(units))
		if err != nil {
			return "", err
		}

		plain.WriteString(string(utf16.Decode(units[next:span.start])))
		plain.WriteString(link.linked(string(utf16.Decode(units[span.start:span.end]))))

		next = span.end
	}

	plain.WriteString(string(utf16.Decode(units[next:])))

	return strings.TrimSpace(plain.String()), nil
}

func linksInOrder(attributes []noteAttribute) []noteAttribute {
	links := slices.DeleteFunc(slices.Clone(attributes), func(attribute noteAttribute) bool {
		return !attribute.isLink()
	})

	slices.SortFunc(links, func(a, b noteAttribute) int {
		return cmp.Compare(locationOf(a), locationOf(b))
	})

	return links
}

func locationOf(attribute noteAttribute) int {
	if len(attribute.Range) == 0 {
		return 0
	}

	return attribute.Range[0]
}
