package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	keyTheme     = "theme"
	keyLanguages = "languages"
	keyClipboard = "copy.clipboard"
)

func parse(raw rawConfig, defaultBindings rawBindings) (Config, error) {
	theme, themeErr := parseEnum(raw.Theme, ParseTheme, ThemeNames())
	languages, languagesErr := parseLanguages(raw.Languages)
	clipboard, clipboardErr := parseEnum(raw.Copy.Clipboard, ParseClipboardBackend, ClipboardBackendNames())
	bindings, bindingsErr := layerBindings(defaultBindings, raw.Bindings)

	return Config{
		Editor:    raw.Editor,
		Theme:     theme,
		Languages: languages,
		Mouse:     raw.Mouse,
		Copy: Copy{
			Clipboard:           clipboard,
			TrimTrailingNewline: raw.Copy.TrimTrailingNewline,
			QuitAfter:           raw.Copy.QuitAfter,
		},
		Bindings: bindings,
	}, errors.Join(atKey(keyTheme, themeErr), languagesErr, atKey(keyClipboard, clipboardErr), bindingsErr)
}

func parseEnum[T any](raw string, parseValue func(string) (T, error), names []string) (T, error) {
	parsed, err := parseValue(raw)
	if err != nil {
		return parsed, fmt.Errorf("%q %w %s", raw, errNotOneOf, strings.Join(names, ", "))
	}

	return parsed, nil
}

func parseLanguages(raw []string) ([]value.Language, error) {
	languages := make([]value.Language, 0, len(raw))
	problems := make([]error, 0, len(raw))

	for i, name := range raw {
		language, err := value.NewLanguage(name)
		if err != nil {
			problems = append(
				problems,
				atKey(fmt.Sprintf("%s[%d]", keyLanguages, i), fmt.Errorf("%q %w", name, errNotALanguage)),
			)

			continue
		}

		languages = append(languages, language)
	}

	return languages, errors.Join(problems...)
}

func atKey(path string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", path, err)
}
