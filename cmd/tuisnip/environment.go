package main

import "io"

type environment struct {
	stdout   io.Writer
	stderr   io.Writer
	getenv   func(key string) string
	environ  func() []string
	lookPath func(file string) (string, error)
	goos     string
}
