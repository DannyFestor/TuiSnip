package tagchoice

import (
	"cmp"
	"slices"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type Chosen struct {
	stored  []domain.Tag
	created []value.TagName
}

func Of(stored []domain.Tag) Chosen {
	return Chosen{stored: slices.SortedFunc(slices.Values(stored), domain.CompareTags), created: nil}
}

func (c Chosen) Stored() []domain.Tag {
	return slices.Clone(c.stored)
}

func (c Chosen) Created() []value.TagName {
	return slices.Clone(c.created)
}

func (c Chosen) Carries(tag domain.Tag) bool {
	return slices.ContainsFunc(c.stored, sameTag(tag))
}

func (c Chosen) CarriesNew(name value.TagName) bool {
	return slices.ContainsFunc(c.created, sameName(name))
}

func (c Chosen) Toggled(tag domain.Tag) Chosen {
	if c.Carries(tag) {
		c.stored = slices.DeleteFunc(slices.Clone(c.stored), sameTag(tag))

		return c
	}

	c.stored = slices.SortedFunc(slices.Values(append(slices.Clone(c.stored), tag)), domain.CompareTags)

	return c
}

func (c Chosen) ToggledNew(name value.TagName) Chosen {
	if c.CarriesNew(name) {
		c.created = slices.DeleteFunc(slices.Clone(c.created), sameName(name))

		return c
	}

	c.created = slices.SortedFunc(slices.Values(append(slices.Clone(c.created), name)), compareNames)

	return c
}

func (c Chosen) Names() []string {
	names := make([]value.TagName, 0, len(c.stored)+len(c.created))
	for _, tag := range c.stored {
		names = append(names, tag.Name())
	}

	names = append(names, c.created...)
	slices.SortStableFunc(names, compareNames)

	spelled := make([]string, 0, len(names))
	for _, name := range names {
		spelled = append(spelled, name.String())
	}

	return spelled
}

func (c Chosen) Equal(other Chosen) bool {
	return slices.EqualFunc(c.stored, other.stored, func(a, b domain.Tag) bool { return a.ID() == b.ID() }) &&
		slices.EqualFunc(c.created, other.created, func(a, b value.TagName) bool { return a.Key() == b.Key() })
}

func sameTag(tag domain.Tag) func(domain.Tag) bool {
	return func(other domain.Tag) bool { return other.ID() == tag.ID() }
}

func sameName(name value.TagName) func(value.TagName) bool {
	return func(other value.TagName) bool { return other.Key() == name.Key() }
}

func compareNames(a, b value.TagName) int {
	return cmp.Compare(a.Key(), b.Key())
}
