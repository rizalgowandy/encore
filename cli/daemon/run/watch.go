package run

import (
	"path/filepath"
	"time"

	"github.com/rjeczalik/notify"
)

// watch watches the given app for changes, and reports
// them on c.
func (mgr *Manager) watch(run *Run) error {
	evs := make(chan notify.EventInfo)
	if err := notify.Watch(filepath.Join(run.App.Root(), "..."), evs, notify.All); err != nil {
		return err
	}

	go func() {
		<-run.Done()
		notify.Stop(evs)
	}()

	go func() {
		for {
			select {
			case <-run.Done():
				return
			case ev := <-evs:
				if ignoreEvent(ev) {
					continue
				}
				// We've seen that some editors like vim rename the .go files to another extension,
				// which breaks our parser since it doesn't recognize the file as a .go file.
				// This race is annoying, but in practice a 100ms delay is imperceptible since
				// the user is busy working in their editor.
				time.Sleep(100 * time.Millisecond)
				mgr.runStdout(run, []byte("Changes detected, recompiling...\n"))
				if err := run.Reload(); err != nil {
					mgr.runStderr(run, []byte(err.Error()))
				} else {
					mgr.runStdout(run, []byte("Reloaded successfully.\n"))
				}
			}
		}
	}()
	return nil
}

func ignoreEvent(ev notify.EventInfo) bool {
	path := ev.Path()

	// Ignore non-Go files
	ext := filepath.Ext(path)
	switch ext {
	case ".go", ".sql", ".mod", ".sum", ".app":
		return false
	default:
		return true
	}
}
