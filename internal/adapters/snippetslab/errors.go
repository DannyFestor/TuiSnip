package snippetslab

import "errors"

var (
	ErrNotAnExport       = errors.New("snippetslab: not a SnippetsLab JSON export")
	ErrMalformedNoteLink = errors.New(
		"snippetslab: a note link's range lies outside its note or overlaps another link",
	)
)
