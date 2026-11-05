package utils

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

func TestAltArrowIsNotCtrl(t *testing.T) {
	var c AltChord
	dir, used := c.Handle(tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModAlt), func(func()) {}, func() {})
	if dir != 1 || !used {
		t.Fatalf("alt+right dir=%d used=%v", dir, used)
	}
	dir, used = c.Handle(tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModMeta), func(func()) {}, func() {})
	if dir != -1 || !used {
		t.Fatalf("meta+left dir=%d used=%v", dir, used)
	}
	dir, used = c.Handle(tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModCtrl), func(func()) {}, func() {})
	if dir != 0 || used {
		t.Fatalf("ctrl+right should be ignored, dir=%d used=%v", dir, used)
	}
}

func TestEscapeThenArrowIsAlt(t *testing.T) {
	var c AltChord
	cancelled := false
	_, used := c.Handle(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone), func(fn func()) { fn() }, func() { cancelled = true })
	if !used || cancelled {
		t.Fatal("leading escape should wait, not cancel")
	}
	dir, used := c.Handle(tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone), func(func()) {}, func() { cancelled = true })
	if dir != -1 || !used || cancelled {
		t.Fatalf("escape+left dir=%d used=%v cancel=%v", dir, used, cancelled)
	}
	time.Sleep(100 * time.Millisecond)
	if cancelled {
		t.Fatal("late escape timer cancelled after the arrow")
	}
}
