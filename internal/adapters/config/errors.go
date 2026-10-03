package config

import "errors"

var ErrFileTooLarge = errors.New("config: file is larger than 1 MiB")

var (
	errNotOneOf       = errors.New("is not one of")
	errNotALanguage   = errors.New("is not a Language")
	errUnknownScope   = errors.New("not a Scope")
	errUnknownBinding = errors.New("not a Binding in")
	errNotAKey        = errors.New("is not a key")
	errNeverMatches   = errors.New("never matches")
	errShiftedSymbol  = errors.New("never matches; write the character it types")
	errForcedQuit     = errors.New("always quits and cannot be bound")
	errTypesText      = errors.New("types text in")
	errAlsoBound      = errors.New("is also bound to")
)
