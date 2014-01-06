package agent

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
)

func (a *AgentIPC) handleRegister(client *IPCClient, seq uint64) error {
	var req registerRequest
	if err := client.dec.Decode(&req); err != nil {
		return fmt.Errorf("decode failed: %v", err)
	}

	var names []string
	for _, file := range req.ConfigPaths {
		f, err = os.Open(path)
		defer f.Close()
		a.agent.RegisterProcess(f)
	}

	// Respond
	header := responseHeader{
		Seq:   seq,
		Error: errToString(nil),
	}
	resp := registerResponse{
		Names: []string{"process_1", "process_2"},
	}
	return client.Send(&header, &resp)
}
