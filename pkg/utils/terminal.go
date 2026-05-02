package utils

import (
	"os"
	"sync"

	"golang.org/x/term"
)

var (
	ttyStateMu     sync.Mutex
	originalTState *term.State
)

// SetTerminalEchoEnabled enables/disables local terminal echo.
// When disabled, stdin is put into raw mode so keystrokes are not echoed or line-buffered.
func SetTerminalEchoEnabled(enabled bool) error {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return nil
	}

	ttyStateMu.Lock()
	defer ttyStateMu.Unlock()

	if enabled {
		if originalTState == nil {
			return nil
		}
		err := term.Restore(fd, originalTState)
		if err == nil {
			originalTState = nil
		}
		return err
	}

	if originalTState == nil {
		state, err := term.GetState(fd)
		if err != nil {
			return err
		}
		originalTState = state
	}
	_, err := term.MakeRaw(fd)
	return err
}
