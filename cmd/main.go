package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/BielosX/sloppy/internal"
)

func main() {
	prog := tea.NewProgram(internal.NewModel())
	if _, err := prog.Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Sloppy failed with an error: %v\n", err)
	}
}
