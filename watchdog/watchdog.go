package watchdog

import (
	"fmt"
	"github.com/appio/watchdog/process"
	"strings"
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
	managed        map[string]chan bool
	pMu            sync.Mutex
	manage         chan int
}

func New() *Watchdog {
	return &Watchdog{
		childProcesses: make(map[string]*process.Process),
		managed:        make(map[string]chan bool, 1),
		manage:         make(chan int),
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
	w.manageProcess(p)

	return nil
}

// Remove a process
func (w *Watchdog) Remove(p *process.Process) error {
	w.pMu.Lock()
	defer w.pMu.Unlock()

	if _, exists := w.childProcesses[p.Name]; !exists {
		return fmt.Errorf("process not found: %s", p.Name)
	}

	w.managed[p.Name] <- true

	delete(w.childProcesses, p.Name)
	delete(w.managed, p.Name)

	return nil
}

// FindByName returns a process for the given name or nil if not found
func (w *Watchdog) FindByName(name string) *process.Process {
	w.pMu.Lock()
	defer w.pMu.Unlock()

	return w.childProcesses[name]
}

// Shutdown stops all running processes ready for safe exit
func (w *Watchdog) Shutdown() error {
	fmt.Println("Watchdog shutting down...")
	for _, proc := range w.childProcesses {
		if proc.IsRunning() {
			fmt.Printf("Stopping process: %s\n", proc.Name)
			proc.Stop()
		}
	}

	return nil
}
func (w *Watchdog) manageProcess(p *process.Process) error {
	w.managed[p.Name] = make(chan bool)

	go func() {
		for {
			select {
			case <-w.managed[p.Name]:
				return

			case out := <-p.OutputChan():
				fmt.Printf("[%s] > %s\n", p.Name, strings.TrimRight(string(out), "\n"))
			}
		}
	}()

	return nil
}
