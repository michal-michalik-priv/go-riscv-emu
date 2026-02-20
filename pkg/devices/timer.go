package devices

import (
	"sync/atomic"
	"time"
)

const (
	// MTIMECMP_OFFSET is the offset for the mtimecmp register in CLINT
	MTIMECMP_OFFSET = 0x4000
	// MTIME_OFFSET is the offset for the mtime register in CLINT
	MTIME_OFFSET = 0xBFF8
	// InterruptMTimer is the bit for Machine Timer Interrupt in mip/mie
	InterruptMTimer = uint32(1 << 7)
)

// TimerDevice implements a simple RISC-V timer (part of CLINT).
type TimerDevice struct {
	mtime            atomic.Uint64
	mtimecmp         atomic.Uint64
	baseAddr         uint32
	size             uint32
	interruptHandler func(uint32, bool)
}

// Initialize sets the base address and size of the timer device.
func (d *TimerDevice) Initialize(baseAddress, size uint32) {
	d.baseAddr = baseAddress
	d.size = size
}

// SetInterruptHandler sets the callback for triggering interrupts.
func (d *TimerDevice) SetInterruptHandler(handler func(uint32, bool)) {
	d.interruptHandler = handler
}

// Start launches the asynchronous timer goroutine.
func (d *TimerDevice) Start() {
	go func() {
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			mtime := d.mtime.Add(1)
			mtimecmp := d.mtimecmp.Load()
			if d.interruptHandler != nil {
				if mtime >= mtimecmp {
					d.interruptHandler(InterruptMTimer, true)
				}
			}
		}
	}()
}

// Read reads a byte from the timer device at the specified address.
func (d *TimerDevice) Read(address uint32) (byte, error) {
	offset := address - d.baseAddr
	var val uint64
	if offset >= MTIMECMP_OFFSET && offset < MTIMECMP_OFFSET+8 {
		val = d.mtimecmp.Load()
		return byte((val >> ((offset - MTIMECMP_OFFSET) * 8)) & 0xFF), nil
	} else if offset >= MTIME_OFFSET && offset < MTIME_OFFSET+8 {
		val = d.mtime.Load()
		return byte((val >> ((offset - MTIME_OFFSET) * 8)) & 0xFF), nil
	}
	return 0, nil
}

// Write writes a byte to the timer device at the specified address.
func (d *TimerDevice) Write(address uint32, value byte) error {
	offset := address - d.baseAddr
	if offset >= MTIMECMP_OFFSET && offset < MTIMECMP_OFFSET+8 {
		shift := (offset - MTIMECMP_OFFSET) * 8
		mask := uint64(0xFF) << shift
		for {
			old := d.mtimecmp.Load()
			newVal := (old & ^mask) | (uint64(value) << shift)
			if d.mtimecmp.CompareAndSwap(old, newVal) {
				// When mtimecmp is written, MTIP should be cleared if mtime < mtimecmp
				if d.interruptHandler != nil && d.mtime.Load() < newVal {
					d.interruptHandler(InterruptMTimer, false)
				}
				break
			}
		}
	} else if offset >= MTIME_OFFSET && offset < MTIME_OFFSET+8 {
		shift := (offset - MTIME_OFFSET) * 8
		mask := uint64(0xFF) << shift
		for {
			old := d.mtime.Load()
			newVal := (old & ^mask) | (uint64(value) << shift)
			if d.mtime.CompareAndSwap(old, newVal) {
				break
			}
		}
	}
	return nil
}

// BaseAddress returns the base address of the timer device.
func (d *TimerDevice) BaseAddress() uint32 {
	return d.baseAddr
}

// Size returns the size of the timer device.
func (d *TimerDevice) Size() uint32 {
	return d.size
}
