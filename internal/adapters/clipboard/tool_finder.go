package clipboard

const (
	darwinGOOS = "darwin"

	waylandDisplayVariable = "WAYLAND_DISPLAY"
	x11DisplayVariable     = "DISPLAY"

	localeAllVariable   = "LC_ALL"
	localeCtypeVariable = "LC_CTYPE"
	langVariable        = "LANG"
	utf8Locale          = "UTF-8"

	nothingToComplainAbout = ""
)

type toolSpec struct {
	name        string
	args        []string
	needsLocale bool
	emptyMarker string
}

type toolFamily struct {
	macOS   toolSpec
	wayland toolSpec
	x11     []toolSpec
}

func copyTools() toolFamily {
	return toolFamily{
		macOS:   toolSpec{name: "pbcopy", args: nil, needsLocale: true, emptyMarker: nothingToComplainAbout},
		wayland: toolSpec{name: "wl-copy", args: nil, needsLocale: false, emptyMarker: nothingToComplainAbout},
		x11: []toolSpec{
			{
				name: "xclip", args: []string{"-in", "-selection", "clipboard"}, needsLocale: false,
				emptyMarker: nothingToComplainAbout,
			},
			{
				name: "xsel", args: []string{"--input", "--clipboard"}, needsLocale: false,
				emptyMarker: nothingToComplainAbout,
			},
		},
	}
}

// wl-paste and xclip exit 1 on an empty clipboard, so only their complaint tells it from a failure.
func readTools() toolFamily {
	return toolFamily{
		macOS: toolSpec{name: "pbpaste", args: nil, needsLocale: true, emptyMarker: nothingToComplainAbout},
		wayland: toolSpec{
			name: "wl-paste", args: []string{"--no-newline"}, needsLocale: false, emptyMarker: "Nothing is copied",
		},
		x11: []toolSpec{
			{
				name: "xclip", args: []string{"-out", "-selection", "clipboard"}, needsLocale: false,
				emptyMarker: "There is no owner for the CLIPBOARD selection",
			},
			{
				name: "xsel", args: []string{"--output", "--clipboard"}, needsLocale: false,
				emptyMarker: nothingToComplainAbout,
			},
		},
	}
}

func findTool(options Options, env environment, family toolFamily) (tool, bool) {
	for _, spec := range family.candidates(options.GOOS, env) {
		path, err := options.LookPath(spec.name)
		if err != nil {
			continue
		}

		return tool{
			name:        spec.name,
			path:        path,
			args:        spec.args,
			env:         spec.environment(env),
			emptyMarker: spec.emptyMarker,
		}, true
	}

	return tool{}, false
}

func (f toolFamily) candidates(goos string, env environment) []toolSpec {
	if goos == darwinGOOS {
		return []toolSpec{f.macOS}
	}

	var candidates []toolSpec

	if env.get(waylandDisplayVariable) != "" {
		candidates = append(candidates, f.wayland)
	}

	if env.get(x11DisplayVariable) != "" {
		candidates = append(candidates, f.x11...)
	}

	return candidates
}

func (s toolSpec) environment(env environment) environment {
	if !s.needsLocale || hasLocale(env) {
		return env
	}

	// pbcopy and pbpaste fall back to the C encoding without a locale and mangle non-ASCII text.
	return env.with(localeCtypeVariable, utf8Locale)
}

func hasLocale(env environment) bool {
	return env.get(localeAllVariable) != "" || env.get(localeCtypeVariable) != "" || env.get(langVariable) != ""
}
