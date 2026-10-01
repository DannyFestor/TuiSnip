package config

import "errors"

var ErrFileTooLarge = errors.New("config: file is larger than 1 MiB")

var (
	errNotOneOf       = errors.New("is not one of")
	errNotALanguage   = errors.New("is not a Language")
	errUnknownScope   = errors.New("not a Scope")
	errUnknownBinding = errors.New("not a Binding in")
)
