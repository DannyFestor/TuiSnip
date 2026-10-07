package look

type palette struct {
	accent          string
	muted           string
	cursorText      string
	unfocusedCursor string
	invalid         string
	currentLine     string
	selection       string
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
		currentLine:     "#262626",
		selection:       "#3D3366",
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
		currentLine:     "#F2F2F2",
		selection:       "#DCD3F7",
		codeStyle:       "github",
	}
}
