package agent

import (
	"fmt"
)

func (a *AgentIPC) handleRegister(client *IPCClient, seq uint64) error {
	var req registerRequest
	if err := client.dec.Decode(&req); err != nil {
		return fmt.Errorf("decode failed: %v", err)
	}

	var names []string
	for _, path := range req.ConfigPaths {
		proc, err := a.agent.RegisterProcess(path)
		if err != nil {
			continue
		}

		names = append(names, proc.Name)
	}

	// Respond
	header := responseHeader{
		Seq:   seq,
		Error: errToString(nil),
	}
	resp := registerResponse{
		Names: names,
	}
	return client.Send(&header, &resp)
}

func (a *AgentIPC) handleStart(client *IPCClient, seq uint64) error {
	var req startRequest
	if err := client.dec.Decode(&req); err != nil {
		return fmt.Errorf("decode failed: %v", err)
	}

	var pids []int
	for _, name := range req.Names {
		proc, err := a.agent.StartProcess(name)
		if err != nil {
			continue
		}

		pids = append(pids, proc.PID())
	}

	// Respond
	header := responseHeader{
		Seq:   seq,
		Error: errToString(nil),
	}
	resp := startResponse{
		Pids: pids,
	}
	return client.Send(&header, &resp)
}
