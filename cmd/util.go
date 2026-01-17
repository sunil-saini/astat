package cmd

import (
	"os"

	"golang.org/x/term"
)

// isInteractive checks if the program is running in an interactive terminal
func isInteractive() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
