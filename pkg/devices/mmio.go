package devices

// MMIODevice represents a memory-mapped I/O device.
type MMIODevice interface {
	Read(address uint32) (byte, error)
	Write(address uint32, value byte) error
	BaseAddress() uint32
	Size() uint32
}

// MMIODevices manages a collection of MMIO devices.
type MMIODevices struct {
	devices []MMIODevice
}

// AddDevice adds a new MMIO device to the collection.
func (mmioDevices *MMIODevices) AddDevice(device MMIODevice) {
	mmioDevices.devices = append(mmioDevices.devices, device)
}

// FindDevice finds the MMIO device that contains the specified address.
func (mmioDevices *MMIODevices) FindDevice(address uint32) MMIODevice {
	for _, device := range mmioDevices.devices {
		if address >= device.BaseAddress() &&
			address < device.BaseAddress()+device.Size() {
			return device
		}
	}
	return nil
}
