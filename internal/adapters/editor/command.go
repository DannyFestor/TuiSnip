package editor

import (
	"slices"
	"strings"
)

const (
	visualVariable = "VISUAL"
	editorVariable = "EDITOR"
)

func fallbackEditors() []string {
	return []string{"nvim", "vim", "vi", "nano"}
}

func resolveCommand(options Options, env []string) (string, bool) {
	for _, named := range []string{options.Configured, lookupEnv(env, visualVariable), lookupEnv(env, editorVariable)} {
		if strings.TrimSpace(named) != "" {
			return named, true
		}
	}

	return firstOnPath(options.LookPath)
}

func firstOnPath(lookPath func(file string) (string, error)) (string, bool) {
	for _, name := range fallbackEditors() {
		if _, err := lookPath(name); err == nil {
			return name, true
		}
	}

	return "", false
}

func lookupEnv(env []string, key string) string {
	prefix := key + "="

	for _, entry := range slices.Backward(env) {
		if value, found := strings.CutPrefix(entry, prefix); found {
			return value
		}
	}

	return ""
}
