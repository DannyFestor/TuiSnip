package editoverlay

import "github.com/DannyFestor/TuiSnip/internal/domain/value"

type entered struct {
	title       string
	description string
	language    value.Language
	content     string
}
