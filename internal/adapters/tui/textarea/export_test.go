package textarea

func CursorLineNumber(m *Model) int {
	return m.cursorLineNumber()
}

func TotalVisualLines(m *Model) int {
	return m.totalVisualLines()
}
