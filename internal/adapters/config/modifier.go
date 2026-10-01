package config

import "strings"

type modifiers uint8

const (
	modifierCtrl modifiers = 1 << iota
	modifierAlt
	modifierShift
	modifierMeta
	modifierHyper
	modifierSuper
)

const modifierSeparator = "+"

func bubbleTeaModifierOrder() []modifiers {
	return []modifiers{modifierCtrl, modifierAlt, modifierShift, modifierMeta, modifierHyper, modifierSuper}
}

func parseModifiers(written string) (modifiers, bool) {
	var parsed modifiers

	for name := range strings.SplitSeq(written, modifierSeparator) {
		modifier, ok := parseModifier(name)
		if !ok || parsed.has(modifier) {
			return 0, false
		}

		parsed |= modifier
	}

	return parsed, true
}

func parseModifier(name string) (modifiers, bool) {
	for _, modifier := range bubbleTeaModifierOrder() {
		if modifier.name() == name {
			return modifier, true
		}
	}

	return 0, false
}

func (m modifiers) has(modifier modifiers) bool {
	return m&modifier != 0
}

func (m modifiers) prefix() string {
	var prefix strings.Builder

	for _, modifier := range bubbleTeaModifierOrder() {
		if m.has(modifier) {
			prefix.WriteString(modifier.name())
			prefix.WriteString(modifierSeparator)
		}
	}

	return prefix.String()
}

func (m modifiers) name() string {
	switch m {
	case modifierCtrl:
		return "ctrl"
	case modifierAlt:
		return "alt"
	case modifierShift:
		return "shift"
	case modifierMeta:
		return "meta"
	case modifierHyper:
		return "hyper"
	case modifierSuper:
		return "super"
	}

	return ""
}
