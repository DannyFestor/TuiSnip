package domain

import "errors"

type FieldError struct {
	Field Field
	Err   error
}

func OnField(field Field, err error) error {
	if err == nil {
		return nil
	}

	return FieldError{Field: field, Err: err}
}

func FieldErrors(err error) []FieldError {
	if err == nil {
		return nil
	}

	if fieldErr, ok := err.(FieldError); ok { //nolint:errorlint // walking the tree by hand to collect every FieldError
		return []FieldError{fieldErr}
	}

	var found []FieldError
	for _, child := range unwrapAll(err) {
		found = append(found, FieldErrors(child)...)
	}

	return found
}

func (e FieldError) Error() string {
	return e.Field.String() + ": " + e.Err.Error()
}

func (e FieldError) Unwrap() error {
	return e.Err
}

func unwrapAll(err error) []error {
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		return joined.Unwrap()
	}

	if child := errors.Unwrap(err); child != nil {
		return []error{child}
	}

	return nil
}
