package overlay

import "charm.land/bubbles/v2/key"

//nolint:iface // iface misses calls through an instantiated generic interface; Stack.Render calls ViewUnder.
type Base[O any] interface {
	Overlay[O]
	ViewUnder(hints []key.Binding) string
}
