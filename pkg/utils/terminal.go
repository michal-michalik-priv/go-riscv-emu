package utils

import (
	"os"
	"sync"

	"golang.org/x/term"
)

var (
	ttyStateMu     sync.Mutex
	originalTState *term.State
	stdinFD        = func() int { return int(os.Stdin.Fd()) }
	isTerminal     = term.IsTerminal
	getTermState   = term.GetState
	makeTermRaw    = term.MakeRaw
	restoreTerm    = term.Restore
)

// SetTerminalEchoEnabled enables/disables local terminal echo.
// When disabled, stdin is put into raw mode so keystrokes are not echoed or line-buffered.
func SetTerminalEchoEnabled(enabled bool) error {
	fd := stdinFD()
	if !isTerminal(fd) {
		return nil
	}

	ttyStateMu.Lock()
	defer ttyStateMu.Unlock()

	if enabled {
		if originalTState == nil {
			return nil
		}
		err := restoreTerm(fd, originalTState)
		if err == nil {
			originalTState = nil
		}
		return err
	}

	if originalTState == nil {
		state, err := getTermState(fd)
		if err != nil {
			return err
		}
		originalTState = state
	}
	_, err := makeTermRaw(fd)
	return err
}
