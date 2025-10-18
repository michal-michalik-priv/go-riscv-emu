package devices

import "fmt"

type DummyTTYDevice struct {
	baseAddress uint32
	size        uint32
}

func (d *DummyTTYDevice) Initialize(baseAddress, size uint32) {
	d.baseAddress = baseAddress
	d.size = size
}

// Read reads a byte from the DummyTTY device at the specified address.
// Always returns 0 as DummyTTY does not support reading.
func (d *DummyTTYDevice) Read(address uint32) (byte, error) {
	if address < d.baseAddress || address >= d.baseAddress+d.size {
		return 0, fmt.Errorf(
			"attempted to read from invalid DummyTTY address %X", address)
	}

	// TODO: Implement reading from keyboard later
	return 0, nil
}

func (d *DummyTTYDevice) Write(address uint32, value byte) error {
	if address < d.baseAddress || address >= d.baseAddress+d.size {
		return fmt.Errorf(
			"attempted to write %X to invalid DummyTTY address %X",
			value, address)
	}

	// For now, just print the character to the console
	fmt.Print(string(value))
	return nil
}

func (d *DummyTTYDevice) BaseAddress() uint32 {
	return d.baseAddress
}

func (d *DummyTTYDevice) Size() uint32 {
	return d.size
}
