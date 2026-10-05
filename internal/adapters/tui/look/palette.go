package look

type palette struct {
	accent          string
	muted           string
	cursorText      string
	unfocusedCursor string
	invalid         string
	codeStyle       string
}

func paletteOf(scheme Scheme) palette {
	if scheme == SchemeLight {
		return lightPalette()
	}

	return darkPalette()
}

func darkPalette() palette {
	return palette{
		accent:          "#7D56F4",
		muted:           "#666666",
		cursorText:      "#FFFFFF",
		unfocusedCursor: "#3A3A3A",
		invalid:         "#FF5F5F",
		codeStyle:       "github-dark",
	}
}

func lightPalette() palette {
	return palette{
		accent:          "#5A3CC8",
		muted:           "#767676",
		cursorText:      "#FFFFFF",
		unfocusedCursor: "#D7D7D7",
		invalid:         "#D70000",
		codeStyle:       "github",
	}
}
