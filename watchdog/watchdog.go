package watchdog

import (
	"fmt"
	"github.com/appio/watchdog/process"
	"sync"
)

// The purpose of this package is to act as a registry for processes
// and proxy commands to processes and respond to events from processes, even
// if the event handling is only logging events.
//
// It is also responsible for sending log output to drain channels.
//
// All exported methods in this package are designed to be interacted with by the `agent` package.
//
// In typical operation this package would only be run once as a daemon per host.

type Watchdog struct {
	childProcesses map[string]*process.Process
	pMu            sync.Mutex
}

func New() *Watchdog {
	return &Watchdog{
		childProcesses: make(map[string]*process.Process),
	}
}

// Add a process
func (w *Watchdog) Add(p *process.Process) error {
	w.pMu.Lock()
	defer w.pMu.Unlock()

	if _, exists := w.childProcesses[p.Name]; exists {
		return fmt.Errorf("process already exists: %s", p.Name)
	}

	w.childProcesses[p.Name] = p

	return nil
}

// Remove a process
func (w *Watchdog) Remove(p *process.Process) error {
	w.pMu.Lock()
	defer w.pMu.Unlock()

	if _, exists := w.childProcesses[p.Name]; !exists {
		return fmt.Errorf("process not found: %s", p.Name)
	}

	delete(w.childProcesses, p.Name)

	return nil
}

// FindByName returns a process for the given name or nil if not found
func (w *Watchdog) FindByName(name string) *process.Process {
	w.pMu.Lock()
	defer w.pMu.Unlock()

	return w.childProcesses[name]
}
