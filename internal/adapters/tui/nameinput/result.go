package nameinput

type Result struct {
	Ending Ending
	Name   string
	Err    error
}

func typing() Result {
	return Result{Ending: Typing, Name: "", Err: nil}
}

func cancelled() Result {
	return Result{Ending: Cancelled, Name: "", Err: nil}
}

func committedAs(name string) Result {
	return Result{Ending: Committed, Name: name, Err: nil}
}

func refusedFor(err error) Result {
	return Result{Ending: Refused, Name: "", Err: err}
}
