package devices

// BusDevice represents a memory-mapped I/O device.
type BusDevice interface {
	Read(address uint32) (byte, error)
	Write(address uint32, value byte) error
	BaseAddress() uint32
	Size() uint32
}

// Bus manages a collection of Bus devices.
type Bus struct {
	devices []BusDevice
}

// AddDevice adds a new Bus device to the collection.
func (Bus *Bus) AddDevice(device BusDevice) {
	Bus.devices = append(Bus.devices, device)
}

// FindDevice finds the Bus device that contains the specified address.
func (Bus *Bus) FindDevice(address uint32) BusDevice {
	for _, device := range Bus.devices {
		if address >= device.BaseAddress() &&
			address < device.BaseAddress()+device.Size() {
			return device
		}
	}
	return nil
}
