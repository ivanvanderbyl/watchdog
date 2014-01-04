package agent

import (
	"github.com/hashicorp/serf/testutil"
	"github.com/mitchellh/cli"
	"testing"
	"time"
)

func TestCommand_implements(t *testing.T) {
	var _ cli.Command = new(Command)
}

func TestCommandRun(t *testing.T) {
	shutdownCh := make(chan struct{})
	defer close(shutdownCh)

	ui := new(cli.MockUi)
	c := &Command{
		ShutdownCh: shutdownCh,
		Ui:         ui,
	}

	args := []string{
		"-rpc-addr", getRPCAddr(),
	}

	resultCh := make(chan int)
	go func() {
		resultCh <- c.Run(args)
	}()

	testutil.Yield()

	// Verify it runs "forever"
	select {
	case <-resultCh:
		t.Fatalf("ended too soon, err: %s", ui.ErrorWriter.String())
	case <-time.After(50 * time.Millisecond):
	}

	// Send a shutdown request
	shutdownCh <- struct{}{}

	select {
	case code := <-resultCh:
		if code != 0 {
			t.Fatalf("bad code: %d", code)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("timeout")
	}
}
