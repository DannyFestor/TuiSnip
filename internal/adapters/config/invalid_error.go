package config

import "strings"

const problemIndent = "\n  "

type InvalidError struct {
	Path     string
	Problems error
}

func (e InvalidError) Error() string {
	return "invalid config " + e.Path + ":" + problemIndent + strings.ReplaceAll(
		e.Problems.Error(),
		"\n",
		problemIndent,
	)
}

func (e InvalidError) Unwrap() error {
	return e.Problems
}
