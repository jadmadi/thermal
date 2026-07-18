package render

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Progress renders an animated progress bar to stderr while a long-running
// operation is in flight. It is only active when stderr is a TTY, so it never
// corrupts --json output or pipes. The bar shows a real percentage driven by
// an atomic counter that the worker goroutine increments per row.
type Progress struct {
	total    int64
	current  atomic.Int64
	label    string
	done     chan struct{}
	started  atomic.Bool
	stopOnce sync.Once
}

// NewProgress returns a Progress bound to a total row count. The bar is not
// shown until Start is called.
func NewProgress(label string, total int64) *Progress {
	return &Progress{total: total, label: label, done: make(chan struct{})}
}

// Start launches the renderer goroutine. No-op (and never starts) if stderr is
// not a TTY, so --json output and pipes are never corrupted.
func (p *Progress) Start() {
	if !stderrIsTerminal() || !p.started.CompareAndSwap(false, true) {
		return
	}
	go p.render()
}

func stderrIsTerminal() bool {
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// Increment advances the row counter by n.
func (p *Progress) Increment(n int64) {
	p.current.Add(n)
}

// Done stops the renderer and clears the bar line.
func (p *Progress) Done() {
	if !p.started.Load() {
		return
	}
	p.stopOnce.Do(func() {
		close(p.done)
		// Give the renderer one final tick to paint 100%.
		time.Sleep(60 * time.Millisecond)
		fmt.Fprint(os.Stderr, "\r"+strings.Repeat(" ", 72)+"\r")
	})
}

func (p *Progress) render() {
	width := 24
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-p.done:
			cur := p.current.Load()
			pct := 100
			if p.total > 0 {
				pct = int(float64(cur) * 100 / float64(p.total))
			}
			if pct > 100 {
				pct = 100
			}
			p.paint(width, pct, cur)
			return
		case <-ticker.C:
			cur := p.current.Load()
			pct := 0
			if p.total > 0 {
				pct = int(float64(cur) * 100 / float64(p.total))
			}
			if pct > 99 {
				pct = 99
			}
			p.paint(width, pct, cur)
		}
	}
}

func (p *Progress) paint(width, pct int, cur int64) {
	filled := width * pct / 100
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	fmt.Fprintf(os.Stderr, "\r  %s  [%s] %3d%%  %s/%s rows",
		p.label, bar, pct, compactInt(cur), compactInt(p.total))
}

func compactInt(n int64) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.0fk", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}
