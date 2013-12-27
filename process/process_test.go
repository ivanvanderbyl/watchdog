package process

import (
	_ "regexp"
	"testing"
	"time"
)

func expect(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

func Test_Process_impl(t *testing.T) {
	proc := NewProcess("echo", "/bin/echo", "-n", "Hello World")
	expect(t, "echo", proc.Name)
	expect(t, "/bin/echo", proc.Command[0])
	expect(t, "stopped", proc.Status())
}

func TestProcesStart(t *testing.T) {
	proc := NewProcess("echo", "/bin/sleep", "0.2")
	proc.Run()
	proc.Start()

	go func() {
		for {
			select {
			case <-time.After(2 * time.Second):
				t.Error("Timed out")
				break
			case event := <-proc.Events:
				switch event {
				case StopEvent:
					t.Logf("Process completed with exit status: %d", proc.LastExitStatus)
				case StartEvent:
					t.Logf("Process started with PID: %d", proc.PID)
				}
				break
			}
		}
	}()

	proc.Wait()
	t.Logf("Done")
}

func TestProcesHasPidAfterStart(t *testing.T) {
	proc := NewProcess("echo", "/bin/sleep", "0.2")
	proc.Run()
	proc.Start()

	if proc.pid == 0 {
		t.Error("proc.PID=0")
	} else {
		t.Logf("proc.PID=%d", proc.PID())
	}

	// go func() {
	// 	for {
	// 		select {
	// 		case <-time.After(2 * time.Second):
	// 			t.Error("Timed out")
	// 			break
	// 		case event := <-proc.Events:
	// 			switch event {
	// 			case StopEvent:
	// 				t.Logf("Process completed with exit status: %d", proc.LastExitStatus)
	// 			case StartEvent:
	// 				t.Logf("Process started with PID: %d", proc.PID)
	// 			}
	// 			break
	// 		}
	// 	}
	// }()

	proc.Wait()
	t.Logf("Done")
}

func TestProcesStop(t *testing.T) {
	proc := NewProcess("echo", "/bin/sleep", "3")
	proc.Run()
	proc.Start()

	for {
		select {
		case <-time.After(2 * time.Second):
			t.Fatal("Timed out")
			return
		case event := <-proc.Events:
			switch event {
			case StopEvent:
				t.Logf("Process completed with exit status: %d", proc.LastExitStatus)
				return
			case StartEvent:
				t.Logf("Process started with PID: %d", proc.PID)
				t.Log("Sending stop")

				<-time.After(1 * time.Second)
				proc.Stop()
			}
		}
	}
	t.Log("Mon loop done")
}
