package picker

import "strings"

func matching(choices []Choice, filter string) []int {
	needle := strings.ToLower(filter)
	prefixed := make([]int, 0, len(choices))
	containing := make([]int, 0)

	for index, choice := range choices {
		text := strings.ToLower(choice.Text)

		switch {
		case strings.HasPrefix(text, needle):
			prefixed = append(prefixed, index)
		case strings.Contains(text, needle):
			containing = append(containing, index)
		}
	}

	return append(prefixed, containing...)
}
