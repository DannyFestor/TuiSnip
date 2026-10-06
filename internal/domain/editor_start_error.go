package domain

type EditorStartError struct {
	Err error
}

func (e EditorStartError) Error() string {
	return "domain: external editor could not start: " + e.Err.Error()
}

func (e EditorStartError) Unwrap() error {
	return e.Err
}
