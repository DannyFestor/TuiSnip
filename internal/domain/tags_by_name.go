package domain

import "slices"

type tagsByName map[string]Tag

func newTagsByName(tags []Tag) tagsByName {
	byName := make(tagsByName, len(tags))
	for _, tag := range tags {
		byName[tag.Name().Key()] = tag
	}

	return byName
}

func (n tagsByName) reused(tags []Tag) []Tag {
	reused := make([]Tag, 0, len(tags))
	for _, tag := range tags {
		reused = append(reused, n.matchOr(tag))
	}

	return reused
}

func (n tagsByName) unmatched(tags []Tag) []Tag {
	return slices.DeleteFunc(slices.Clone(tags), n.matches)
}

func (n tagsByName) matchOr(tag Tag) Tag {
	match, found := n[tag.Name().Key()]
	if !found {
		return tag
	}

	return match
}

func (n tagsByName) matches(tag Tag) bool {
	_, found := n[tag.Name().Key()]

	return found
}
