package telemetry

import (
	"testing"
	"time"
)

func TestWindowBounds(t *testing.T) {
	from, to := window()

	ft, err := time.Parse(time.RFC3339, from)
	if err != nil {
		t.Fatalf("from %q not RFC3339: %v", from, err)
	}
	tt, err := time.Parse(time.RFC3339, to)
	if err != nil {
		t.Fatalf("to %q not RFC3339: %v", to, err)
	}
	if !ft.Before(tt) {
		t.Fatalf("from %s should be before to %s", from, to)
	}
	if got := tt.Sub(ft); got < lookback-time.Minute || got > lookback+time.Minute {
		t.Fatalf("window span %s not ~%s", got, lookback)
	}
}
