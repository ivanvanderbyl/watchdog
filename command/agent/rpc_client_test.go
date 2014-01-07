package agent

import (
	"github.com/hashicorp/serf/testutil"
	"io"
	"io/ioutil"
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

var basicConfig string = `{
  "name": "my_app",
  "disabled": false,
  "program": "/usr/local/bin/node",
  "program_arguments": [],
  "keep_alive": false,
  "run_at_load": true
}`

func TestClientRegister(t *testing.T) {
	tf, err := ioutil.TempFile("", "my_app.json")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Write([]byte(basicConfig))
	tf.Close()

	client, agent, ipc := testRPCClient(t)
	defer ipc.Shutdown()
	defer client.Close()
	defer agent.Shutdown()

	if err := agent.Start(); err != nil {
		t.Fatalf("err: %s", err)
	}

	testutil.Yield()

	resp, err := client.Register([]string{tf.Name()}, true, true)
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(resp) == 0 {
		t.Fatal("No processes were registered")
	}

	if resp[0] != "my_app" {
		t.Errorf("Expected process name to be %s, got %v", "my_app", resp[0])
	}
}
