package testutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBenchmarkTimer(t *testing.T) {
	t.Run("Elapsed returns stored time when stopped", func(t *testing.T) {
		// Create a mock benchmark
		b := &testing.B{}
		timer := NewBenchmarkTimer(b)

		// Let some time pass
		time.Sleep(50 * time.Millisecond)

		// Stop the timer
		timer.Stop()
		elapsedWhenStopped := timer.Elapsed()

		// Verify elapsed time is not zero and reasonable
		assert.Greater(t, elapsedWhenStopped, time.Duration(0), "Elapsed time should be greater than 0 when stopped")
		assert.Greater(t, elapsedWhenStopped, 40*time.Millisecond, "Elapsed time should be at least 40ms")
		assert.Less(t, elapsedWhenStopped, 100*time.Millisecond, "Elapsed time should be less than 100ms")

		// Wait a bit more
		time.Sleep(50 * time.Millisecond)

		// Elapsed should still return the same stored value
		elapsedAfterWait := timer.Elapsed()
		assert.Equal(t, elapsedWhenStopped, elapsedAfterWait, "Elapsed time should not change after stop")
	})

	t.Run("Elapsed resets when timer restarts", func(t *testing.T) {
		b := &testing.B{}
		timer := NewBenchmarkTimer(b)

		// Let some time pass
		time.Sleep(50 * time.Millisecond)
		timer.Stop()
		
		// Restart the timer
		timer.Start()
		
		// Immediately check elapsed - should be very small
		elapsed := timer.Elapsed()
		assert.Less(t, elapsed, 10*time.Millisecond, "Elapsed time should be reset after restart")
	})
}