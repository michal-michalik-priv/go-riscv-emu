package utils

import (
	"errors"
	"testing"

	"golang.org/x/term"
)

func withTerminalMocks(t *testing.T) {
	t.Helper()

	origStdinFD := stdinFD
	origIsTerminal := isTerminal
	origGetTermState := getTermState
	origMakeTermRaw := makeTermRaw
	origRestoreTerm := restoreTerm
	origState := originalTState

	t.Cleanup(func() {
		stdinFD = origStdinFD
		isTerminal = origIsTerminal
		getTermState = origGetTermState
		makeTermRaw = origMakeTermRaw
		restoreTerm = origRestoreTerm
		originalTState = origState
	})

	originalTState = nil
}

func TestSetTerminalEchoEnabled_NonTerminal_NoOp(t *testing.T) {
	withTerminalMocks(t)

	stdinFD = func() int { return 99 }
	isTerminal = func(int) bool { return false }
	getTermState = func(int) (*term.State, error) {
		t.Fatal("getTermState should not be called")
		return nil, nil
	}

	if err := SetTerminalEchoEnabled(false); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestSetTerminalEchoEnabled_Disable_SavesStateAndMakesRaw(t *testing.T) {
	withTerminalMocks(t)

	state := &term.State{}
	fd := 10
	calledGet := false
	calledRaw := false

	stdinFD = func() int { return fd }
	isTerminal = func(got int) bool { return got == fd }
	getTermState = func(got int) (*term.State, error) {
		calledGet = true
		if got != fd {
			t.Fatalf("unexpected fd: %d", got)
		}
		return state, nil
	}
	makeTermRaw = func(got int) (*term.State, error) {
		calledRaw = true
		if got != fd {
			t.Fatalf("unexpected fd: %d", got)
		}
		return &term.State{}, nil
	}

	if err := SetTerminalEchoEnabled(false); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !calledGet {
		t.Fatal("expected getTermState to be called")
	}
	if !calledRaw {
		t.Fatal("expected makeTermRaw to be called")
	}
	if originalTState != state {
		t.Fatal("expected original terminal state to be cached")
	}
}

func TestSetTerminalEchoEnabled_Disable_GetStateError(t *testing.T) {
	withTerminalMocks(t)

	expectedErr := errors.New("get state failed")
	stdinFD = func() int { return 7 }
	isTerminal = func(int) bool { return true }
	getTermState = func(int) (*term.State, error) { return nil, expectedErr }
	makeTermRaw = func(int) (*term.State, error) {
		t.Fatal("makeTermRaw should not be called")
		return nil, nil
	}

	err := SetTerminalEchoEnabled(false)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
	if originalTState != nil {
		t.Fatal("expected original state to remain nil on get state error")
	}
}

func TestSetTerminalEchoEnabled_Disable_MakeRawError(t *testing.T) {
	withTerminalMocks(t)

	state := &term.State{}
	expectedErr := errors.New("make raw failed")
	stdinFD = func() int { return 5 }
	isTerminal = func(int) bool { return true }
	getTermState = func(int) (*term.State, error) { return state, nil }
	makeTermRaw = func(int) (*term.State, error) { return nil, expectedErr }

	err := SetTerminalEchoEnabled(false)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
	if originalTState != state {
		t.Fatal("expected original state to stay cached on make raw error")
	}
}

func TestSetTerminalEchoEnabled_Enable_NoSavedState_NoOp(t *testing.T) {
	withTerminalMocks(t)

	stdinFD = func() int { return 3 }
	isTerminal = func(int) bool { return true }
	restoreTerm = func(int, *term.State) error {
		t.Fatal("restoreTerm should not be called")
		return nil
	}

	if err := SetTerminalEchoEnabled(true); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestSetTerminalEchoEnabled_Enable_RestoreSuccess_ClearsState(t *testing.T) {
	withTerminalMocks(t)

	fd := 11
	state := &term.State{}
	originalTState = state
	calledRestore := false

	stdinFD = func() int { return fd }
	isTerminal = func(int) bool { return true }
	restoreTerm = func(gotFD int, gotState *term.State) error {
		calledRestore = true
		if gotFD != fd {
			t.Fatalf("unexpected fd: %d", gotFD)
		}
		if gotState != state {
			t.Fatal("unexpected state pointer")
		}
		return nil
	}

	if err := SetTerminalEchoEnabled(true); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !calledRestore {
		t.Fatal("expected restoreTerm to be called")
	}
	if originalTState != nil {
		t.Fatal("expected cached state to be cleared after successful restore")
	}
}

func TestSetTerminalEchoEnabled_Enable_RestoreError_KeepsState(t *testing.T) {
	withTerminalMocks(t)

	state := &term.State{}
	originalTState = state
	expectedErr := errors.New("restore failed")

	stdinFD = func() int { return 13 }
	isTerminal = func(int) bool { return true }
	restoreTerm = func(int, *term.State) error { return expectedErr }

	err := SetTerminalEchoEnabled(true)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
	if originalTState != state {
		t.Fatal("expected cached state to remain on restore failure")
	}
}
