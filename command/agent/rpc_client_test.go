package agent

import (
	"github.com/hashicorp/serf/testutil"
	"io"
	"net"
	"os"
	"testing"
)

func TestClientHandshake(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	lw := NewLogWriter(512)
	mult := io.MultiWriter(os.Stderr, lw)

	agent := testAgent(mult)
	ipc := NewAgentIPC(agent, l, mult, lw)

	client, err := NewRPCClient(l.Addr().String())
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	defer ipc.Shutdown()
	defer client.Close()
	defer agent.Shutdown()

	if err := agent.Start(); err != nil {
		t.Fatalf("err: %s", err)
	}

	testutil.Yield()

}
