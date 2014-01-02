package agent

import (
	"github.com/appio/watchdog/watchdog"
	"sync"
)

// Agent starts and manages the Watchdog instance.
type Agent struct {
	dog *watchdog.Watchdog

	// shutdownCh is used for shutdowns
	shutdown     bool
	shutdownCh   chan struct{}
	shutdownLock sync.Mutex
}

func NewAgent() *Agent {
	return &Agent{
		dog:        watchdog.New(),
		shutdownCh: make(chan struct{}),
	}
}

func (a *Agent) Start() error {
	return nil
}

func (a *Agent) Shutdown() error {
	return nil
}

// ShutdownCh returns a channel that can be selected to wait
// for the agent to perform a shutdown.
func (a *Agent) ShutdownCh() <-chan struct{} {
	return a.shutdownCh
}
