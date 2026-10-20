package textarea

type drawnRows struct {
	first int
	end   int
}

func (d drawnRows) overlaps(first, count int) bool {
	return first < d.end && first+count > d.first
}
