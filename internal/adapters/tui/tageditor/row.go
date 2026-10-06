package tageditor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/picker"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	carriedMark   = "✓ "
	uncarriedMark = "  "
	createMark    = "+ "
	newTagMeta    = "new"
	createPrefix  = `create "`
	createSuffix  = `"`
)

type row interface {
	key() string
	choice(state toggles) picker.Choice
	toggledIn(state toggles) (toggles, error)
	landing() string
}

type storedRow struct {
	counted browse.TagCount
}

func (r storedRow) key() string {
	return r.counted.Tag.Name().Key()
}

func (r storedRow) choice(state toggles) picker.Choice {
	return picker.Choice{
		Mark:     r.mark(state),
		Text:     r.landing(),
		Meta:     strconv.Itoa(r.counted.SnippetCount),
		Trailing: false,
	}
}

func (r storedRow) mark(state toggles) string {
	if state.chosen.Carries(r.counted.Tag) {
		return carriedMark
	}

	return uncarriedMark
}

func (r storedRow) toggledIn(state toggles) (toggles, error) {
	state.chosen = state.chosen.Toggled(r.counted.Tag)

	return state, nil
}

func (r storedRow) landing() string {
	return r.counted.Tag.Name().String()
}

type newRow struct {
	name value.TagName
}

func (r newRow) key() string {
	return r.name.Key()
}

func (r newRow) choice(state toggles) picker.Choice {
	return picker.Choice{Mark: r.mark(state), Text: r.landing(), Meta: newTagMeta, Trailing: false}
}

func (r newRow) mark(state toggles) string {
	if state.chosen.CarriesNew(r.name) {
		return carriedMark
	}

	return uncarriedMark
}

func (r newRow) toggledIn(state toggles) (toggles, error) {
	state.chosen = state.chosen.ToggledNew(r.name)

	return state, nil
}

func (r newRow) landing() string {
	return r.name.String()
}

type createRow struct {
	typed string
}

func (r createRow) key() string {
	return value.TagNameKey(r.typed)
}

func (r createRow) choice(toggles) picker.Choice {
	return picker.Choice{
		Mark:     createMark,
		Text:     createPrefix + r.landing() + createSuffix,
		Meta:     newTagMeta,
		Trailing: true,
	}
}

func (r createRow) toggledIn(state toggles) (toggles, error) {
	name, err := value.NewTagName(r.typed)
	if err != nil {
		return state, fmt.Errorf("tageditor: %w", err)
	}

	return state.withCreated(name), nil
}

func (r createRow) landing() string {
	return strings.TrimSpace(r.typed)
}
