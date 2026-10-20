package runeutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/textarea/runeutil"
)

func TestSanitize(t *testing.T) {
	t.Parallel()

	td := []struct {
		input, output string
	}{
		{"", ""},
		{"x", "x"},
		{"\n", "XX"},
		{"\na\n", "XXaXX"},
		{"\n\n", "XXXX"},
		{"\t", "\t"},
		{"hello", "hello"},
		{"hel\nlo", "helXXlo"},
		{"hel\rlo", "helXXlo"},
		{"hel\tlo", "hel\tlo"},
		{"he\n\nl\tlo", "heXXXXl\tlo"},
		{"he\tl\n\nlo", "he\tlXXXXlo"},
		{"hel\x1blo", "hello"},
		{"hello\xc2", "hello"}, // invalid utf8
	}

	for _, tc := range td {
		sanitizer := runeutil.NewSanitizer(runeutil.ReplaceNewlines("XX"))

		result := sanitizer.Sanitize([]rune(tc.input))

		assert.Equal(t, tc.output, string(result), "input %q", tc.input)
	}
}
