// Package textarea is TuiSnip's fork of the Bubbles v2.2.1 textarea, the
// Content textarea of the edit overlay and Capture. TuiSnip owns it because
// the stock textarea can't keep tabs or highlight while editing (ADR 0004).
// It keeps the upstream editor and nothing TuiSnip-specific: the edit overlay
// decides what may be typed or pasted, and input maps look.Styles onto it.
package textarea
