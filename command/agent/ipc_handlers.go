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
