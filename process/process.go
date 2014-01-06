package process

import (
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
	"time"
)

type ProcessState int
type CtrlState int
type Event int

const (
	ProcessStopped ProcessState = iota
	ProcessStarting
	ProcessRunning
	ProcessStopping
)

const (
	StartEvent Event = iota
	StopEvent
)

const (
	COMMAND_START int = iota
	COMMAND_STOP
	COMMAND_RESTART
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

func (e *Event) String() string {
	switch *e {
	case StartEvent:
		return "start"
	case StopEvent:
		return "stop"
	}
	return "unknown"
}

// Process represents a process running under command of Nord agent
type Process struct {
	// Process Name
	Name string `json:"name"`

	// Enabled indicates whether to run this process. A value of false will ensure
	// it is always stopped
	Enabled bool `json:"enabled"`

	// Process Identifier
	pid int `json:"pid,int"`

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
	KeepAlive bool `json:"keep_alive"`

	// User and Group to switch to after exec
	UserName  string `json:"user"`
	GroupName string `json:"group"`

	// WorkingDirectory is the directory to chdir to after forking
	WorkingDirectory string `json:"working_directory"`

	// Outlets are used to send the process output to other services
	outlets []*io.Writer

	// Internal state of the process
	state ProcessState

	proc       *os.Process
	outputChan chan []byte
	done       chan int
	Events     chan Event
	manage     chan *processCommand
	waitChan   chan bool

	runner ProcessRunner

	stateMu sync.Mutex
	pidMu   sync.Mutex

	sync.Mutex
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
		Throttle:    time.Second * 10,

		outputChan: make(chan []byte),
		done:       make(chan int),
		manage:     make(chan *processCommand),
		Events:     make(chan Event),
		waitChan:   make(chan bool),
	}
}

// NewProcessFromConfig creates a process from a ProcessConf
func NewProcessFromConfig(conf *ProcessConfig) *Process {

	return &Process{}
}

func (p *Process) OutputChan() chan []byte {
	return p.outputChan
}

// Status returns the process state
func (p *Process) Status() string {
	p.stateMu.Lock()
	defer p.stateMu.Unlock()
	return p.state.String()
}

func (p *Process) setStatus(state ProcessState) {
	p.stateMu.Lock()
	defer p.stateMu.Unlock()
	p.state = state
}

func (p *Process) Wait() {
	<-p.waitChan
}

// SetRunner sets the process runner (useful for testing with a mock runner)
// FIXME(ivanvanderbyl): This probably doesn't need to be exported
func (p *Process) SetRunner(r ProcessRunner) {
	p.runner = r
}

// Start the process
func (p *Process) Start() error {
	replyChan := make(chan error)
	c := &processCommand{COMMAND_START, replyChan}
	p.manage <- c
	return <-c.Reply
}

func (p *Process) Stop() error {
	replyChan := make(chan error)
	c := &processCommand{COMMAND_STOP, replyChan}
	p.manage <- c
	return <-c.Reply
}

func (p *Process) Restart() error {
	replyChan := make(chan error)
	c := &processCommand{COMMAND_RESTART, replyChan}
	p.manage <- c
	return <-c.Reply
}

func (p *Process) exec() error {
	p.Lock()
	defer p.Unlock()

	if p.runner == nil {
		p.runner = &DefaultRunner{}
	}

	p.setStatus(ProcessStarting)
	p.StartedAt = time.Now()

	proc, err := p.runner.Exec(p, p.outputChan, p.done)
	if err != nil {
		return err
	}

	p.proc = proc

	p.setStatus(ProcessRunning)
	return nil
}

func (p *Process) setPid(pid int) {
	p.pidMu.Lock()
	defer p.pidMu.Unlock()
	p.pid = pid
}

func (p *Process) PID() int {
	p.pidMu.Lock()
	defer p.pidMu.Unlock()
	return p.pid
}

func (p *Process) terminate() error {
	// didExit := make(chan bool)

	err := p.proc.Signal(syscall.SIGQUIT)
	if err != nil {
		return err
	}

	if p.PID() == 0 {
	}
	return nil

	// return p.proc.Kill()
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
	go p.runloop()
}

func (p *Process) runloop() {
	go func(p *Process) {
		// var willExit bool = false

		for {
			select {
			case status := <-p.done:
				fmt.Println("Processs Exited")
				p.finish(status)

				select {
				case p.waitChan <- true:
				default:
				}

				select {
				case p.Events <- StopEvent:
				default:
				}

				// 				if p.KeepAlive && !willExit {
				// 					<-time.After(p.Throttle)
				// 					p.Start()
				// 				}

				// willExit = false

			case command := <-p.manage:
				switch command.Command {
				case COMMAND_START:
					err := p.exec()
					if err != nil {
						fmt.Println("Failed to start process:", err.Error())
					} else {
						// p.Events <- StartEvent
						select {
						case p.Events <- StartEvent:
						default:
							fmt.Println("Start Event Ignored")
						}
					}
					command.Reply <- err

					fmt.Println("Started")

				case COMMAND_STOP:
					fmt.Println("Received stop command", p.proc.Pid)
					if p.proc != nil {
						p.terminate()
					} else {
						fmt.Println("proc is nil")
					}
					command.Reply <- nil

					// err := p.terminate()
					// if err != nil {
					// 	fmt.Println("Error terminating process", err.Error())
					// }

				case COMMAND_RESTART:
					if p.PID() != 0 {
						p.Stop()
						<-time.After(p.Throttle)
					}
					p.Start()
					command.Reply <- nil
				}

				// 			case command := <-p.commands:
				// 				switch command.Command {
				// 				case COMMAND_START:
				// 					command.Reply <- p.exec()

				// 				case COMMAND_STOP:
				// 					// User initiated stop, do not relaunch
				// 					willExit = true
				// 					command.Reply <- p.terminate()
				// 				}
			}
		}
	}(p)
}
