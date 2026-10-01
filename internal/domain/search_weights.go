package domain

const (
	titleSearchWeight       = 4
	tagSearchWeight         = 3
	descriptionSearchWeight = 2
	contentSearchWeight     = 1
)

func bestWeightedScore(scores FieldScores) int {
	best := max(
		scores.Title*titleSearchWeight,
		scores.Description*descriptionSearchWeight,
		scores.Content*contentSearchWeight,
	)
	for _, tagScore := range scores.Tags {
		best = max(best, tagScore*tagSearchWeight)
	}

	return best
}
