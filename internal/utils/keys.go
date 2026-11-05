package utils

import (
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

// AltChord turns Alt+Left and Alt+Right into a direction.
// Typical terminals only deliver that chord when Shift is held too
// (Alt+Shift+arrow). Plain Alt+arrow is often stolen, or sent as Escape
// followed by an arrow, which used to cancel the dialog.
type AltChord struct {
	mu    sync.Mutex
	armed bool
}

// HasAlt reports whether Alt or Meta is held. Terminals disagree on the name.
func HasAlt(m tcell.ModMask) bool {
	return m&tcell.ModAlt != 0 || m&tcell.ModMeta != 0
}

// Handle returns dir -1 or 1 when the user pressed Alt+Left or Alt+Right.
// consumed means the key should not be passed on.
// later must run fn on the UI goroutine, for example app.QueueUpdateDraw.
// It is only called from a background goroutine, never from the key handler.
func (a *AltChord) Handle(event *tcell.EventKey, later func(func()), cancel func()) (dir int, consumed bool) {
	if event == nil {
		return 0, false
	}
	a.mu.Lock()
	if a.armed {
		a.armed = false
		a.mu.Unlock()
		switch event.Key() {
		case tcell.KeyLeft:
			return -1, true
		case tcell.KeyRight:
			return 1, true
		case tcell.KeyEscape:
			cancel()
			return 0, true
		default:
			cancel()
			return 0, false
		}
	}

	if HasAlt(event.Modifiers()) {
		a.mu.Unlock()
		switch event.Key() {
		case tcell.KeyLeft:
			return -1, true
		case tcell.KeyRight:
			return 1, true
		default:
			return 0, false
		}
	}

	if event.Key() == tcell.KeyEscape && event.Modifiers() == tcell.ModNone {
		a.armed = true
		a.mu.Unlock()
		go func() {
			time.Sleep(60 * time.Millisecond)
			later(func() {
				a.mu.Lock()
				still := a.armed
				a.armed = false
				a.mu.Unlock()
				if still {
					cancel()
				}
			})
		}()
		return 0, true
	}
	a.mu.Unlock()
	return 0, false
}
