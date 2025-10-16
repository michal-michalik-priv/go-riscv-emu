package memory

import (
	"testing"
)

func setupRAMFixture(t *testing.T) *RandomAccessMemory {
	t.Helper()

	ramSize := uint32(1024) // 1 KB RAM
	ram := NewRAM(ramSize)

	return ram
}

func TestRAMReadWrite(t *testing.T) {
	ram := setupRAMFixture(t)

	// Test writing and reading a byte within bounds
	address := uint32(100)
	value := byte(42)
	ram.Write(address, value)
	readValue, err := ram.Read(address)
	if err != nil {
		t.Errorf("Error reading RAM: %v", err)
		return
	}
	if readValue != value {
		t.Errorf("Expected value %d at address %d, got %d", value, address,
			readValue)
	}
}

func TestRAMOutOfBounds(t *testing.T) {
	ram := setupRAMFixture(t)

	// Test reading out of bounds
	_, err := ram.Read(ram.Size())
	if err == nil {
		t.Errorf("Expected error when reading out of bounds, got nil")
	}

	// Test writing out of bounds
	err = ram.Write(ram.Size(), 42)
	if err == nil {
		t.Errorf("Expected error when writing out of bounds, got nil")
	}
}
