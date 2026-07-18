package render

import (
	"sync"
	"testing"
)

func TestProgress_ConcurrencySafe(t *testing.T) {
	p := NewProgress("Test", 1000)
	// Even if not TTY, calling Increment, Start, and Done concurrently must not race or panic
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.Start()
			p.Increment(100)
			p.Done()
			p.Done() // Double call must not panic
		}()
	}
	wg.Wait()
	if p.current.Load() != 1000 {
		t.Errorf("expected current 1000, got %d", p.current.Load())
	}
}
