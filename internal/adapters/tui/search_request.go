package tui

type searchRequest int

const (
	searchStays searchRequest = iota
	searchQueries
	searchReveals
	searchCopies
	searchCloses
)
