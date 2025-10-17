package devices

import "testing"

type MockBusDevice struct {
	baseAddress uint32
	size        uint32
	memory      []byte
}

func (m *MockBusDevice) Read(address uint32) (byte, error) {
	offset := address - m.baseAddress
	return m.memory[offset], nil
}

func (m *MockBusDevice) Write(address uint32, value byte) error {
	offset := address - m.baseAddress
	m.memory[offset] = value
	return nil
}

func (m *MockBusDevice) BaseAddress() uint32 {
	return m.baseAddress
}

func (m *MockBusDevice) Size() uint32 {
	return m.size
}

func (m *MockBusDevice) Initialize(baseAddress, size uint32) {
	m.baseAddress = baseAddress
	m.size = size
	m.memory = make([]byte, size)
}

func setupBusFixture() *Bus {
	bus := &Bus{}

	// Create and add a mock Bus device
	mockDevice := &MockBusDevice{}
	mockDevice.Initialize(0x1000, 256) // Base address 0x1000, size 256 bytes
	bus.AddDevice(mockDevice)

	return bus
}

func TestBusReadWrite(t *testing.T) {
	bus := setupBusFixture()

	// Test writing and reading a byte within the mock device
	address := uint32(0x1000) // Within the mock device range
	value := byte(55)
	device := bus.FindDevice(address)
	if device == nil {
		t.Errorf("Expected to find device at address %X, got nil", address)
		return
	}

	err := device.Write(address, value)
	if err != nil {
		t.Errorf("Error writing to Bus device: %v", err)
		return
	}

	readValue, err := device.Read(address)
	if err != nil {
		t.Errorf("Error reading from Bus device: %v", err)
		return
	}

	if readValue != value {
		t.Errorf("Expected value %d at address %X, got %d", value, address,
			readValue)
	}
}

func TestBusOutOfBounds(t *testing.T) {
	bus := setupBusFixture()

	// Test reading out of bounds
	address := uint32(0x2000) // Outside the mock device range
	device := bus.FindDevice(address)
	if device != nil {
		t.Errorf("Expected no device at address %X, but found one", address)
	}
}
