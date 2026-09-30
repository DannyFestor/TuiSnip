// PROTOTYPE — throwaway answer to "Main screen layout and navigation" (issue #16).
// Three variants of the main screen, switchable with ` and ~ or -variant=A|B|C.
// Fake in-memory data; nothing is saved, copied, or written.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func main() {
	start := flag.String("variant", "A", "variant to start on: A, B, or C")
	empty := flag.Bool("empty", false, "start with no Snippets or Folders, to see the empty-screen hint")
	flag.Parse()

	variants := []variant{threeColumns{}, stacked{}, focusWidens{}}
	idx := int(strings.ToUpper(*start)[0] - 'A')
	if idx < 0 || idx >= len(variants) {
		idx = 0
	}
	if _, err := tea.NewProgram(newModel(seed(*empty), variants, idx)).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
