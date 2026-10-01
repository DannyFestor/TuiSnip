package value

import "strings"

type SearchQuery struct{ value string }

func NewSearchQuery(raw string) SearchQuery {
	return SearchQuery{value: raw}
}

func (q SearchQuery) IsBlank() bool {
	return strings.TrimSpace(q.value) == ""
}

func (q SearchQuery) String() string {
	return q.value
}
