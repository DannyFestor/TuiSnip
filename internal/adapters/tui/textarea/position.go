package textarea

type Position struct {
	Row int
	Col int
}

func (p Position) before(q Position) bool {
	if p.Row != q.Row {
		return p.Row < q.Row
	}

	return p.Col < q.Col
}
