package overlay

//nolint:iface // iface misses calls through an instantiated generic interface; round.bubble calls Received.
type Parent[O any] interface {
	Overlay[O]
	Received(outcome O) Step[O]
}
