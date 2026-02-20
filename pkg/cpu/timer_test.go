package cpu

import (
	"testing"
	"time"

	"github.com/michal-michalik-priv/go-riscv-emu/pkg/devices"
)

func TestTimerInterrupt(t *testing.T) {
	bus := &devices.Bus{}
	core := NewCore(bus)
	core.mode = ModeMachine

	timer := &devices.TimerDevice{}
	timer.Initialize(0x02000000, 0x10000)
	timer.SetInterruptHandler(core.SetInterrupt)
	bus.AddDevice(timer)

	// Set mtimecmp to a small value
	// mtimecmp is at offset 0x4000
	// We'll write 10 to mtimecmp
	timer.Write(0x02004000, 10)

	// Start the timer
	timer.Start()

	// Wait for mtime to reach mtimecmp
	// Since timer increments every 1ms, it should take ~10ms
	timeout := time.After(100 * time.Millisecond)
	ticker := time.NewTicker(1 * time.Millisecond)
	defer ticker.Stop()

	found := false
	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for timer interrupt")
		case <-ticker.C:
			core.mu.Lock()
			if (core.csrs[csrMip] & (1 << 7)) != 0 {
				found = true
			}
			core.mu.Unlock()
		}
		if found {
			break
		}
	}
}

func TestWFIWakeup(t *testing.T) {
	bus := &devices.Bus{}
	core := NewCore(bus)
	core.mode = ModeMachine

	// Set WFI state
	core.wfi = true

	// In a separate goroutine, trigger an interrupt after 50ms
	go func() {
		time.Sleep(50 * time.Millisecond)
		core.SetInterrupt(1<<7, true) // MTIP
	}()

	// Try to Step. Should busy-wait until interrupt.
	// We simulate the main loop.
	timeout := time.After(200 * time.Millisecond)
	start := time.Now()
	for {
		err := Step(core)
		if err != nil {
			t.Fatalf("Step failed: %v", err)
		}

		core.mu.Lock()
		wfi := core.wfi
		core.mu.Unlock()

		if !wfi {
			break
		}

		select {
		case <-timeout:
			t.Fatal("Timeout waiting for WFI wakeup")
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}

	elapsed := time.Since(start)
	if elapsed < 40*time.Millisecond {
		t.Errorf("WFI woke up too early: %v", elapsed)
	}
}
