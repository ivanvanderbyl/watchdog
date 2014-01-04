package agent

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
)

func getRPCAddr() string {
	for i := 0; i < 500; i++ {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", rand.Int31n(25000)+1024))
		if err == nil {
			l.Close()
			return l.Addr().String()
		}
	}

	panic("no listener")
}

func testAgent(logOutput io.Writer) *Agent {
	if logOutput == nil {
		logOutput = os.Stderr
	}
	config := DefaultConfig

	agent := NewAgent(config, logOutput)

	return agent
}
