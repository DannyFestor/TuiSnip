package domain

import "strconv"

type FragmentError struct {
	Position int
	Err      error
}

func InFragment(position int, err error) error {
	if err == nil {
		return nil
	}

	return FragmentError{Position: position, Err: err}
}

func (e FragmentError) Error() string {
	return "fragment " + strconv.Itoa(e.Position+1) + ": " + e.Err.Error()
}

func (e FragmentError) Unwrap() error {
	return e.Err
}
