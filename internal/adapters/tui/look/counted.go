package look

import "fmt"

type Tally struct {
	Count int
	Noun  string
}

func Counted(count int, noun string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, noun)
	}

	return fmt.Sprintf("%d %ss", count, noun)
}

func NonZeroCounts(tallies ...Tally) []string {
	var counts []string

	for _, tally := range tallies {
		if tally.Count > 0 {
			counts = append(counts, Counted(tally.Count, tally.Noun))
		}
	}

	return counts
}
