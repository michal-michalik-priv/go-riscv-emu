package devices

import (
	"testing"
)

func TestDummyTTYDevice_Write(t *testing.T) {
	tty := DummyTTYDevice{}
	var baseAddress uint32 = 0x2000
	var size uint32 = 0x10
	tty.Initialize(baseAddress, size)

	// Test writing a byte to the Dummy TTY device
	address := baseAddress + 0x5
	value := byte('A')

	err := tty.Write(address, value)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
}

func TestDummyTTYDevice_Read(t *testing.T) {
	tty := DummyTTYDevice{}
	var baseAddress uint32 = 0x2000
	var size uint32 = 0x10
	tty.Initialize(baseAddress, size)

	// Test reading a byte from the Dummy TTY device
	address := baseAddress + 0x5

	readValue, err := tty.Read(address)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// TODO: Currently, DummyTTY Read always returns 0
	if readValue != 0 {
		t.Errorf("Expected read value 0, got %X", readValue)
	}
}

func TestDummyTTYDevice_OutOfBounds(t *testing.T) {
	tty := DummyTTYDevice{}
	var baseAddress uint32 = 0x2000
	var size uint32 = 0x10
	tty.Initialize(baseAddress, size)

	// Test write out of bounds
	err := tty.Write(baseAddress+size, 'B')
	if err == nil {
		t.Error("Expected error for out-of-bounds write, got nil")
	}

	// Test read out of bounds
	_, err = tty.Read(baseAddress + size)
	if err == nil {
		t.Error("Expected error for out-of-bounds read, got nil")
	}

	// Test read below base address
	_, err = tty.Read(baseAddress - 1)
	if err == nil {
		t.Error("Expected error for read below base address, got nil")
	}

	// Test write below base address
	err = tty.Write(baseAddress-1, 'C')
	if err == nil {
		t.Error("Expected error for write below base address, got nil")
	}
}

func TestDummyTTYDevice_BaseAddress(t *testing.T) {
	tty := DummyTTYDevice{}
	var baseAddress uint32 = 0x2000
	var size uint32 = 0x10
	tty.Initialize(baseAddress, size)

	if tty.BaseAddress() != baseAddress {
		t.Errorf("Expected BaseAddress %X, got %X", baseAddress,
			tty.BaseAddress())
	}
}

func TestDummyTTYDevice_Size(t *testing.T) {
	tty := DummyTTYDevice{}
	var baseAddress uint32 = 0x2000
	var size uint32 = 0x10
	tty.Initialize(baseAddress, size)

	if tty.Size() != size {
		t.Errorf("Expected Size %X, got %X", size, tty.Size())
	}
}
