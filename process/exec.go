package process

import (
	"os"
	"os/exec"
	"syscall"
)

// outputWriter is a simple io.Writer which writes to a channel
type outputWriter struct {
	out chan<- []byte
}

func (w *outputWriter) Write(b []byte) (n int, err error) {
	w.out <- b
	return len(b), nil
}

type DefaultRunner struct{}

// Exec launches the given process
func (r *DefaultRunner) Exec(p *Process, outputChan chan []byte, done chan int) (proc *os.Process, err error) {
	var exitStatus int = 0

	executable, err := exec.LookPath(p.Command[0])
	if err != nil {
		return proc, err
	}

	cmd := exec.Command(executable, p.Command[1:]...)
	cmd.Env = append(cmd.Env, p.formattedEnv()...)

	writer := &outputWriter{outputChan}

	cmd.Stderr = writer
	cmd.Stdout = writer

	if err := cmd.Start(); err != nil {
		return proc, err
	}

	go func() {
		defer func() {
			p.Lock()
			defer p.Unlock()
			p.setPid(0)
		}()

		// Wait for the process to exit
		err := cmd.Wait()
		if err != nil {
			switch err.(type) {
			case *exec.ExitError:
				exitStatus = err.(*exec.ExitError).Sys().(syscall.WaitStatus).ExitStatus()
			case *os.PathError:
				exitStatus = 127
			}
			// fmt.Println("Process Exited (wait complete)", exitStatus, err.Error())
		}
		done <- exitStatus
	}()

	return cmd.Process, nil
}
