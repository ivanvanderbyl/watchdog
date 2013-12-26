package process

import (
	"testing"
	"time"
)

func TestExec(t *testing.T) {
	proc := NewProcess("echo", "/bin/echo", "-n", "Hello World")

	outChan := make(chan []byte, 1)
	statusChan := make(chan int, 1)

	runner := &DefaultRunner{}
	go runner.Exec(proc, outChan, statusChan)

	select {
	case out := <-outChan:
		if string(out) == "Hello World" {
			t.Logf("Received correct output")
		} else {
			t.Errorf("Unexpected output: %s", string(out))
		}

	case <-time.After(1 * time.Second):
		t.Error("Timed out waiting for output")
	}

	select {
	case status := <-statusChan:
		if status == 0 {
			t.Logf("Process exited with status %d", status)
		} else {
			t.Errorf("Process exited with status %d", status)
		}
	case <-time.After(1 * time.Second):
		t.Error("Exec timed out")
	}
}
