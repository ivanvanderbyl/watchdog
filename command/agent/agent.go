package agent

import (
	"github.com/appio/watchdog/process"
	"github.com/appio/watchdog/watchdog"
	"io"
	"log"
	"os"
	"sync"
)

// Agent starts and manages the Watchdog instance.
type Agent struct {
	// logger instance wraps the logOutput
	logger *log.Logger

	dog *watchdog.Watchdog

	// shutdownCh is used for shutdowns
	shutdown     bool
	shutdownCh   chan struct{}
	shutdownLock sync.Mutex
}

func NewAgent(config *Config, logOutput io.Writer) *Agent {
	// Ensure we have a log sink
	if logOutput == nil {
		logOutput = os.Stderr
	}

	return &Agent{
		dog:        watchdog.New(),
		logger:     log.New(logOutput, "", log.LstdFlags),
		shutdownCh: make(chan struct{}),
	}
}

func (a *Agent) Start() error {
	a.logger.Println("[INFO] Watchdog starting...")

	return nil
}

func (a *Agent) Shutdown() error {
	a.logger.Println("[INFO] Gracefully shutting down...")
	return a.dog.Shutdown()
}

// ShutdownCh returns a channel that can be selected to wait
// for the agent to perform a shutdown.
func (a *Agent) ShutdownCh() <-chan struct{} {
	return a.shutdownCh
}

// RegisterProcess takes a configuration file and registers a new process
func (a *Agent) RegisterProcess(configPath string) (*process.Process, error) {
	config, err := process.LoadConfigFile(configPath)
	if err != nil {
		return nil, err
	}

	proc := process.NewProcessFromConfig(config)
	a.dog.Add(proc)

	a.logger.Printf("[INFO] Registered process: %s", config.Name)

	return proc, nil
}
