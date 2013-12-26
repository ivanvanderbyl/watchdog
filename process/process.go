package process

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
)

type ProcessState int

const (
	ProcessStopped ProcessState = iota
	ProcessStarting
	ProcessRunning
	ProcessStopping
)

const (
	COMMAND_START int = iota
	COMMAND_STOP
)

type processCommand struct {
	Command int
	Reply   chan error
}

func (p *ProcessState) String() string {
	switch *p {
	case ProcessStopped:
		return "stopped"
	case ProcessStarting:
		return "starting"
	case ProcessRunning:
		return "running"
	case ProcessStopping:
		return "stopping"
	}
	return "unknown"
}

// Process represents a process running under command of Nord agent
type Process struct {
	// Process Name
	Name string `json:"name"`

	// Process Identifier
	PID int `json:"pid,int"`

	// Last status code from this process exiting
	LastExitStatus int `json:"last_exit_status,int"`

	// Launch timeout
	Timeout time.Duration `json:"timeout"`

	// Command is the executable and arguments to run
	Command []string `json:"command"`

	// Environment is a hash of environment vars to set for this process
	Environment map[string]string `json:"environment"`

	// The time this process started
	StartedAt time.Time `json:"started_at"`

	// Signal to send to the process to gracefully exit
	KillSignal os.Signal `json:"kill_signal"`

	// Timeout to wait for process to exit gracefully before killing
	KillTimeout time.Duration `json:"kill_timeout"`

	// Throttle relaunching
	Throttle time.Duration `json:"throttle"`

	// Restart process it exits
	KeepAlive bool

	// User and Group to switch to after exec
	User  string `json:"user"`
	Group string `json:"group"`

	// Internal state of the process
	state ProcessState

	proc       os.Process
	outputChan chan []byte
	done       chan int
	commands   chan processCommand

	runner ProcessRunner

	stateMu *sync.Mutex

	*sync.Mutex
}

// ProcessRunner is an interface for running processes, used mainly for switching
// between a live runner and a test runner
type ProcessRunner interface {
	Exec(*Process, chan []byte, chan int) (*os.Process, error)
}

// NewProcess constructs a new Process instance which can be accepted by
// the management queue
func NewProcess(name string, command ...string) *Process {
	return &Process{
		Name:        name,
		Command:     command,
		state:       ProcessStopped,
		Environment: make(map[string]string),
		KillTimeout: time.Second * 10,
		KillSignal:  os.Signal(syscall.SIGQUIT),
		KeepAlive:   true,
		Throttle:    10 * time.Millisecond,

		outputChan: make(chan []byte, 0),
		done:       make(chan int),
		commands:   make(chan processCommand, 1),

		stateMu: new(sync.Mutex),
		Mutex:   new(sync.Mutex),
	}
}

func (p *Process) OutputChan() chan []byte {
	return p.outputChan
}

func (p *Process) Status() string {
	p.stateMu.Lock()
	defer p.stateMu.Unlock()
	return p.state.String()
}

// SetRunner sets the process runner (useful for testing with a mock runner)
// FIXME(ivanvanderbyl): This probably doesn't need to be exported
func (p *Process) SetRunner(r ProcessRunner) {
	p.runner = r
}

func (p *Process) exec() error {
	p.Lock()
	defer p.Unlock()

	if p.runner == nil {
		p.runner = &DefaultRunner{}
	}

	p.StartedAt = time.Now()

	proc, err := p.runner.Exec(p, p.outputChan, p.done)
	if err != nil {
		return err
	}

	p.state = ProcessRunning

	p.setPid(proc.Pid)
	return nil
}

func (p *Process) setPid(pid int) {
	p.PID = pid
}

func (p *Process) terminate() error {
	return p.proc.Kill()
}

func (p *Process) formattedEnv() []string {
	var env []string

	for k, v := range p.Environment {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return env
}

func (p *Process) finish(status int) {
	p.stateMu.Lock()
	defer p.stateMu.Unlock()

	p.StartedAt = time.Time{}
	p.LastExitStatus = status
	p.state = ProcessStopped
}

func (p *Process) Run() {
	// go p.runloop()
}

// func (p *Process) runloop() {
// 	go func(p *Process) {
// 		var willExit bool = false

// 		for {
// 			select {
// 			case status := <-p.done:
// 				fmt.Println("Processs Exited")

// 				p.finish(status)

// 				if p.KeepAlive && !willExit {
// 					<-time.After(p.Throttle)
// 					p.Start()
// 				}

// 				willExit = false

// 			case command := <-p.commands:
// 				switch command.Command {
// 				case COMMAND_START:
// 					command.Reply <- p.exec()

// 				case COMMAND_STOP:
// 					// User initiated stop, do not relaunch
// 					willExit = true
// 					command.Reply <- p.terminate()
// 				}
// 			}
// 		}
// 	}(p)
// }
