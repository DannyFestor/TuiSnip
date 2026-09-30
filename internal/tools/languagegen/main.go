package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

const generatedFileMode = 0o644

var errMissingOutPath = errors.New("languagegen: -out is required")

func main() {
	outPath := flag.String("out", "", "path of the Go file to write")

	flag.Parse()

	if err := run(*outPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(outPath string) error {
	if outPath == "" {
		return errMissingOutPath
	}

	source, err := renderSource(languageNames())
	if err != nil {
		return err
	}

	if err := os.WriteFile(outPath, source, generatedFileMode); err != nil {
		return fmt.Errorf("languagegen: %w", err)
	}

	return nil
}
