package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/BielosX/sloppy/internal"
)

func main() {
	var err error
	bedrock, err := internal.NewBedrockClient()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to initialize AWS Bedrock Client: %v\n", err)
		os.Exit(1)
	}
	prog := tea.NewProgram(internal.NewModel(bedrock))
	if _, err = prog.Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Sloppy failed with an error: %v\n", err)
		os.Exit(1)
	}
}
