package worker

import (
	"testing"
	"testing/synctest"
	"time"
)

func TestBackoff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		n := 10 * time.Minute
		d := n

		// 10min -> 20 -> 40 -> 60 -> 60 ...
		for i := range 5 {
			backoff(&d, 1*time.Hour)

			if i > 2 {
				if d != 1*time.Hour {
					t.Errorf("cap: got %v, want 1h", d)
				}
			} else {
				x := n * 2 * time.Duration(i+1)
				if d != x {
					t.Errorf("next wait: got %v, want %v", d, x)
				}
			}
		}
	})
}
