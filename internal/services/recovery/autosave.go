package recovery

import (
	"os"
	"sync"
	"sync/atomic"
	"time"

	"gosheet/internal/services/fileop"

	"github.com/rivo/tview"
)

type Snapshot struct {
	Sheets      []fileop.SheetInfo
	ActiveSheet int
}

type CaptureFunc func() (Snapshot, bool)

// Autosaver captures workbook state on tview's UI thread and writes the
// detached snapshot in the background.
type Autosaver struct {
	app      *tview.Application
	path     string
	interval time.Duration
	capture  CaptureFunc
	onError  func(error)

	mu         sync.Mutex
	generation uint64
	stopped    atomic.Bool
	pending    atomic.Bool
	stop       chan struct{}
	done       chan struct{}
	stopOnce   sync.Once
	workers    sync.WaitGroup
}

func NewAutosaver(app *tview.Application, path string, interval time.Duration, capture CaptureFunc, onError func(error)) *Autosaver {
	a := &Autosaver{
		app:      app,
		path:     path,
		interval: interval,
		capture:  capture,
		onError:  onError,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
	if interval <= 0 {
		a.interval = 5 * time.Minute
	}
	go a.run()
	return a
}

func (a *Autosaver) run() {
	defer close(a.done)
	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()
	for {
		select {
		case <-a.stop:
			return
		case <-ticker.C:
			a.queueCapture()
		}
	}
}

func (a *Autosaver) queueCapture() {
	if a.stopped.Load() || !a.pending.CompareAndSwap(false, true) {
		return
	}
	go func() {
		a.app.QueueUpdateDraw(func() {
			defer a.pending.Store(false)
			if a.stopped.Load() {
				return
			}
			snapshot, dirty := a.capture()
			if !dirty {
				a.report(a.Clear())
				return
			}
			a.persist(snapshot)
		})
	}()
}

func (a *Autosaver) persist(snapshot Snapshot) {
	a.mu.Lock()
	if a.stopped.Load() {
		a.mu.Unlock()
		return
	}
	generation := a.generation
	a.workers.Add(1)
	a.mu.Unlock()

	go func() {
		defer a.workers.Done()
		a.mu.Lock()
		if generation != a.generation {
			a.mu.Unlock()
			return
		}
		err := fileop.SaveRecoveryWorkbook(a.path, snapshot.Sheets, snapshot.ActiveSheet)
		a.mu.Unlock()
		a.report(err)
	}()
}

// Clear invalidates queued snapshots before removing the last recovery file.
func (a *Autosaver) Clear() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.generation++
	err := os.Remove(a.path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (a *Autosaver) Stop() {
	a.stopOnce.Do(func() {
		a.mu.Lock()
		a.stopped.Store(true)
		close(a.stop)
		a.mu.Unlock()
	})
	<-a.done
	a.workers.Wait()
}

func (a *Autosaver) report(err error) {
	if err != nil && a.onError != nil {
		a.onError(err)
	}
}
