package agent

import (
	"fmt"
)

func (a *AgentIPC) handleRegister(client *IPCClient, seq uint64) error {
	var req registerRequest
	if err := client.dec.Decode(&req); err != nil {
		return fmt.Errorf("decode failed: %v", err)
	}

	// a.logger.Printf("Got Register Request: %v\n", req)

	num := len(req.ConfigPaths)

	// a.agent.dog.Add(p)

	// Respond
	header := responseHeader{
		Seq:   seq,
		Error: errToString(nil),
	}
	resp := registerResponse{
		Num: int32(num),
	}
	return client.Send(&header, &resp)
}
