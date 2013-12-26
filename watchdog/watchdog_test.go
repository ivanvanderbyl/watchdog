package watchdog

import (
	"github.com/appio/watchdog/process"
	"testing"
)

func TestAddProcess(t *testing.T) {
	watchdog := New()

	p := process.NewProcess("echo", "/bin/echo")
	watchdog.Add(p)

	if len(watchdog.childProcesses) != 1 {
		t.Error("Failed to add process")
	}
}

func TestRemoveProcess(t *testing.T) {
	watchdog := New()

	p := process.NewProcess("echo", "/bin/echo")
	watchdog.Add(p)

	if len(watchdog.childProcesses) != 1 {
		t.Error("Failed to add process")
	}

	watchdog.Remove(p)

	if len(watchdog.childProcesses) != 0 {
		t.Error("Failed to remove process")
	}
}

func TestFindProcessByName(t *testing.T) {
	watchdog := New()
	p1 := process.NewProcess("echo", "/bin/echo")
	watchdog.Add(p1)

	p2 := watchdog.FindByName("echo")
	if p2 == nil {
		t.Fatal("failed to find process by name")
	}

	if p1 != p2 {
		t.Fatal("expected found process to be equal to first process")
	}
}
