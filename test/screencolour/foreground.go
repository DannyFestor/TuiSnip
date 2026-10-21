package screencolour

import (
	"regexp"
	"strings"
	"testing"
)

const (
	parameterSeparator = ";"
	resetCode          = "0"
	defaultForeground  = "39"
	foregroundCode     = "38"
	backgroundCode     = "48"
	rgbMode            = "2"
	rgbParameters      = 5
)

var escapeParameters = regexp.MustCompile(`\x1b\[([0-9;]*)m`)

func Foreground(tb testing.TB, screen, text string) string {
	tb.Helper()

	shown := regexp.MustCompile(`((?:\x1b\[[0-9;]*m)*)` + regexp.QuoteMeta(text)).FindStringSubmatch(screen)
	if shown == nil {
		tb.Fatalf("%q is not on the screen", text)

		return ""
	}

	foreground := ""
	for _, escape := range escapeParameters.FindAllStringSubmatch(shown[1], -1) {
		foreground = foregroundIn(strings.Split(escape[1], parameterSeparator), foreground)
	}

	return foreground
}

func foregroundIn(parameters []string, current string) string {
	for index := 0; index < len(parameters); index++ {
		code := parameters[index]

		switch {
		case code == "" || code == resetCode || code == defaultForeground:
			current = ""
		case isRGB(parameters, index):
			if code == foregroundCode {
				current = strings.Join(parameters[index+2:index+rgbParameters], parameterSeparator)
			}

			index += rgbParameters - 1
		}
	}

	return current
}

func isRGB(parameters []string, index int) bool {
	code := parameters[index]

	return (code == foregroundCode || code == backgroundCode) &&
		index+rgbParameters <= len(parameters) && parameters[index+1] == rgbMode
}
