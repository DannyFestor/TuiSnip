package snippetslab

import "fmt"

const (
	linkAttributeName       = "link"
	locationAndLengthFields = 2
)

type noteAttribute struct {
	Name        string `json:"name"`
	Range       []int  `json:"range"`
	StringValue string `json:"stringValue"`
}

type unitSpan struct {
	start int
	end   int
}

func (a noteAttribute) isLink() bool {
	return a.Name == linkAttributeName
}

func (a noteAttribute) span(from, noteLength int) (unitSpan, error) {
	if len(a.Range) != locationAndLengthFields {
		return unitSpan{}, fmt.Errorf("range %v: %w", a.Range, ErrMalformedNoteLink)
	}

	span := unitSpan{start: a.Range[0], end: a.Range[0] + a.Range[1]}
	if span.start < from || span.end < span.start || span.end > noteLength {
		return unitSpan{}, fmt.Errorf("range %v: %w", a.Range, ErrMalformedNoteLink)
	}

	return span, nil
}

func (a noteAttribute) linked(text string) string {
	if text == a.StringValue {
		return text
	}

	return text + " (" + a.StringValue + ")"
}
