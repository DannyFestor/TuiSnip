package tui

type editRequest int

const (
	editStays editRequest = iota
	editSaves
	editCancels
	editRefusesPaste
)
