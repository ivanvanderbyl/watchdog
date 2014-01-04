package command

import (
	"flag"
	"github.com/appio/watchdog/command/agent"
)

// RPCAddrFlag returns a pointer to a string that will be populated
// when the given flagset is parsed with the RPC address of the Watchdog.
func RPCAddrFlag(f *flag.FlagSet) *string {
	return f.String("rpc-addr", "127.0.0.1:6673",
		"RPC address of the Watchdog agent")
}

// RPCClient returns a new Serf RPC client with the given address.
func RPCClient(addr string) (*agent.RPCClient, error) {
	return agent.NewRPCClient(addr)
}
