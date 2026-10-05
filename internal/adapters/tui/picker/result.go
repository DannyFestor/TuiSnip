package picker

type Result struct {
	Ending Ending
	Index  int
}

func filtering() Result {
	return Result{Ending: Filtering, Index: 0}
}

func cancelled() Result {
	return Result{Ending: Cancelled, Index: 0}
}

func pickedAt(index int) Result {
	return Result{Ending: Picked, Index: index}
}
