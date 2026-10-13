package domain

const (
	titleSearchWeight         = 4
	tagSearchWeight           = 3
	fragmentTitleSearchWeight = 3
	descriptionSearchWeight   = 2
	contentSearchWeight       = 1
)

func bestWeightedSnippetFieldScore(scores FieldScores) int {
	best := max(scores.Title*titleSearchWeight, scores.Description*descriptionSearchWeight)
	for _, tagScore := range scores.Tags {
		best = max(best, tagScore*tagSearchWeight)
	}

	return best
}

func weightedFragmentScore(scores FragmentScores) int {
	return max(scores.Title*fragmentTitleSearchWeight, scores.Content*contentSearchWeight)
}
