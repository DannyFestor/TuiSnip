package clipboard

const (
	darwinGOOS = "darwin"

	waylandDisplayVariable = "WAYLAND_DISPLAY"
	x11DisplayVariable     = "DISPLAY"

	localeAllVariable   = "LC_ALL"
	localeCtypeVariable = "LC_CTYPE"
	langVariable        = "LANG"
	utf8Locale          = "UTF-8"
)

type toolSpec struct {
	name        string
	args        []string
	needsLocale bool
}

func findTool(options Options, env environment) (tool, bool) {
	for _, spec := range toolCandidates(options.GOOS, env) {
		path, err := options.LookPath(spec.name)
		if err != nil {
			continue
		}

		return tool{name: spec.name, path: path, args: spec.args, env: spec.environment(env)}, true
	}

	return tool{}, false
}

func toolCandidates(goos string, env environment) []toolSpec {
	if goos == darwinGOOS {
		return []toolSpec{{name: "pbcopy", args: nil, needsLocale: true}}
	}

	var candidates []toolSpec

	if env.get(waylandDisplayVariable) != "" {
		candidates = append(candidates, toolSpec{name: "wl-copy", args: nil, needsLocale: false})
	}

	if env.get(x11DisplayVariable) != "" {
		candidates = append(candidates,
			toolSpec{name: "xclip", args: []string{"-in", "-selection", "clipboard"}, needsLocale: false},
			toolSpec{name: "xsel", args: []string{"--input", "--clipboard"}, needsLocale: false},
		)
	}

	return candidates
}

func (s toolSpec) environment(env environment) environment {
	if !s.needsLocale || hasLocale(env) {
		return env
	}

	// pbcopy falls back to the C encoding without a locale and mangles non-ASCII text.
	return env.with(localeCtypeVariable, utf8Locale)
}

func hasLocale(env environment) bool {
	return env.get(localeAllVariable) != "" || env.get(localeCtypeVariable) != "" || env.get(langVariable) != ""
}
