package folder

type MoveRepository interface {
	Finder
	DescendantLister
	Mover
}
