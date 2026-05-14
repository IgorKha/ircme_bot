package metrics

import "testing"

func TestCounterInc(t *testing.T) {
	t.Parallel()

	counter := NewCounter()
	if got := counter.Inc(); got != 1 {
		t.Fatalf("Inc() = %d, want 1", got)
	}

	if got := counter.Inc(); got != 2 {
		t.Fatalf("Inc() = %d, want 2", got)
	}

	if got := counter.Value(); got != 2 {
		t.Fatalf("Value() = %d, want 2", got)
	}
}
