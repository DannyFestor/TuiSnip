// Package mainscreen is the main screen: the four Panes, which one has focus,
// how the layout shares the terminal between them, and the status line. It sits
// at the bottom of the Overlay stack, opens the Overlays its Bindings ask for,
// and reports the rest as outcomes, without running an Action itself.
package mainscreen
