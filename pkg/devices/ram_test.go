package devices

import (
	"testing"
)

// setupRAMDevice creates and initializes a RAMDevice for testing.
func setupRAMDevice() (*RAMDevice, uint32, uint32) {
	ram := &RAMDevice{}
	var baseAddress uint32 = 0x1000
	var size uint32 = 0x100
	ram.Initialize(baseAddress, size)
	return ram, baseAddress, size
}

func TestRAMDevice_Initialize(t *testing.T) {
	ram, baseAddress, size := setupRAMDevice()

	if ram.baseAddress != baseAddress {
		t.Errorf("Expected baseAddress %X, got %X", baseAddress, ram.baseAddress)
	}
	if ram.size != size {
		t.Errorf("Expected size %X, got %X", size, ram.size)
	}
	if ram.memory == nil {
		t.Error("RAM memory not initialized")
	}
	if ram.memory.Size() != size {
		t.Errorf("Expected internal memory size %X, got %X", size, ram.memory.Size())
	}
}

func TestRAMDevice_ReadWrite(t *testing.T) {
	ram, baseAddress, _ := setupRAMDevice()

	// Test valid write and read
	var testAddress uint32 = baseAddress + 0x10
	var testValue byte = 0xAB

	err := ram.Write(testAddress, testValue)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	readValue, err := ram.Read(testAddress)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if readValue != testValue {
		t.Errorf("Expected value %X, got %X", testValue, readValue)
	}
}

func TestRAMDevice_OutOfBounds(t *testing.T) {
	ram, baseAddress, size := setupRAMDevice()

	// Test write out of bounds
	err := ram.Write(baseAddress+size, 0xFF)
	if err == nil {
		t.Error("Expected error for out-of-bounds write, got nil")
	}

	// Test read out of bounds
	_, err = ram.Read(baseAddress + size)
	if err == nil {
		t.Error("Expected error for out-of-bounds read, got nil")
	}

	// Test read below base address
	_, err = ram.Read(baseAddress - 1)
	if err == nil {
		t.Error("Expected error for read below base address, got nil")
	}

	// Test write below base address
	err = ram.Write(baseAddress-1, 0xFF)
	if err == nil {
		t.Error("Expected error for write below base address, got nil")
	}
}

func TestRAMDevice_BaseAddress(t *testing.T) {
	ram, baseAddress, _ := setupRAMDevice()

	if ram.BaseAddress() != baseAddress {
		t.Errorf("Expected BaseAddress %X, got %X", baseAddress,
			ram.BaseAddress())
	}
}

func TestRAMDevice_Size(t *testing.T) {
	ram, _, size := setupRAMDevice()

	if ram.Size() != size {
		t.Errorf("Expected Size %X, got %X", size, ram.Size())
	}
}
