package agent

import (
	"github.com/hashicorp/serf/testutil"
	"io"
	"net"
	"os"
	"testing"
)

func testRPCClient(t *testing.T) (*RPCClient, *Agent, *AgentIPC) {
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

	return client, agent, ipc
}

func TestClientRegister(t *testing.T) {
	client, agent, ipc := testRPCClient(t)
	defer ipc.Shutdown()
	defer client.Close()
	defer agent.Shutdown()

	if err := agent.Start(); err != nil {
		t.Fatalf("err: %s", err)
	}

	testutil.Yield()

	resp, err := client.Register([]string{"/etc/watchdog.json"}, true, true)
	if err != nil {
		t.Fatal(err.Error())
	}

	if resp[0] != "process_1" {
		t.Errorf("Expected process name to be %s", "process_1")
	}
}
