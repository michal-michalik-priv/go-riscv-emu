package cpu

import (
	"testing"
)

func TestSetPc(t *testing.T) {
	core := NewCore()
	newPc := uint32(0x1000)
	core.SetPc(newPc)

	if core.pc != newPc {
		t.Errorf("Expected PC to be %X, got %X", newPc, core.pc)
	}
}
