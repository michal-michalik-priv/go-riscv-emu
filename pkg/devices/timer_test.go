package devices

import (
	"sync/atomic"
	"testing"
	"time"
)

func writeUint64ToTimer(t *testing.T, d *TimerDevice, startAddr uint32, value uint64) {
	t.Helper()
	for i := uint32(0); i < 8; i++ {
		err := d.Write(startAddr+i, byte((value>>(i*8))&0xFF))
		if err != nil {
			t.Fatalf("write failed at offset %d: %v", i, err)
		}
	}
}

func readUint64FromTimer(t *testing.T, d *TimerDevice, startAddr uint32) uint64 {
	t.Helper()
	var v uint64
	for i := uint32(0); i < 8; i++ {
		b, err := d.Read(startAddr + i)
		if err != nil {
			t.Fatalf("read failed at offset %d: %v", i, err)
		}
		v |= uint64(b) << (i * 8)
	}
	return v
}

func TestTimerDevice_InitializeAndMetadata(t *testing.T) {
	d := &TimerDevice{}
	d.Initialize(0x02000000, 0x10000)

	if d.BaseAddress() != 0x02000000 {
		t.Fatalf("expected base address 0x02000000, got %X", d.BaseAddress())
	}
	if d.Size() != 0x10000 {
		t.Fatalf("expected size 0x10000, got %X", d.Size())
	}
}

func TestTimerDevice_ReadWriteRegisters(t *testing.T) {
	d := &TimerDevice{}
	d.Initialize(0x02000000, 0x10000)

	expectedMtime := uint64(0x1122334455667788)
	expectedMtimecmp := uint64(0x8877665544332211)

	writeUint64ToTimer(t, d, 0x02000000+MTIME_OFFSET, expectedMtime)
	writeUint64ToTimer(t, d, 0x02000000+MTIMECMP_OFFSET, expectedMtimecmp)

	gotMtime := readUint64FromTimer(t, d, 0x02000000+MTIME_OFFSET)
	if gotMtime != expectedMtime {
		t.Fatalf("expected mtime %X, got %X", expectedMtime, gotMtime)
	}

	gotMtimecmp := readUint64FromTimer(t, d, 0x02000000+MTIMECMP_OFFSET)
	if gotMtimecmp != expectedMtimecmp {
		t.Fatalf("expected mtimecmp %X, got %X", expectedMtimecmp, gotMtimecmp)
	}

	if d.GetMtime() != expectedMtime {
		t.Fatalf("GetMtime expected %X, got %X", expectedMtime, d.GetMtime())
	}
	if d.GetMtimecmp() != expectedMtimecmp {
		t.Fatalf("GetMtimecmp expected %X, got %X", expectedMtimecmp, d.GetMtimecmp())
	}
}

func TestTimerDevice_ReadUnknownOffsetReturnsZero(t *testing.T) {
	d := &TimerDevice{}
	d.Initialize(0x02000000, 0x10000)

	b, err := d.Read(0x02000000 + 0x1234)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if b != 0 {
		t.Fatalf("expected zero for unknown offset, got %X", b)
	}
}

func TestTimerDevice_WriteMtimecmpClearsInterrupt(t *testing.T) {
	d := &TimerDevice{}
	d.Initialize(0x02000000, 0x10000)

	var sawClear atomic.Bool
	d.SetInterruptHandler(func(interrupt uint32, active bool) {
		if interrupt == InterruptMTimer && !active {
			sawClear.Store(true)
		}
	})

	writeUint64ToTimer(t, d, 0x02000000+MTIME_OFFSET, 5)

	if err := d.Write(0x02000000+MTIMECMP_OFFSET, 10); err != nil {
		t.Fatalf("write mtimecmp failed: %v", err)
	}

	if !sawClear.Load() {
		t.Fatalf("expected mtimecmp write to clear MTIP when mtime < mtimecmp")
	}
}

func TestTimerDevice_StartRaisesTimerInterrupt(t *testing.T) {
	d := &TimerDevice{}
	d.Initialize(0x02000000, 0x10000)

	events := make(chan bool, 16)
	d.SetInterruptHandler(func(interrupt uint32, active bool) {
		if interrupt != InterruptMTimer {
			return
		}
		select {
		case events <- active:
		default:
		}
	})

	if err := d.Write(0x02000000+MTIMECMP_OFFSET, 1); err != nil {
		t.Fatalf("write mtimecmp failed: %v", err)
	}

	d.Start()

	timeout := time.After(100 * time.Millisecond)
	for {
		select {
		case active := <-events:
			if active {
				return
			}
		case <-timeout:
			t.Fatal("timeout waiting for active timer interrupt")
		}
	}
}
