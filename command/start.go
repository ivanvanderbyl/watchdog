package command

import (
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"strings"
)

// StartCommand starts a registered process
type StartCommand struct {
	Ui cli.Ui
}

func (c *StartCommand) Help() string {
	helpText := `
Usage: watchdog start [options] <process_name> ...

  Starts a process.

Options:

  -rpc-addr=127.0.0.1:6673  RPC address of the Watchdog agent.
`
	return strings.TrimSpace(helpText)
}

func (c *StartCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("start", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	rpcAddr := RPCAddrFlag(cmdFlags)
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	processNames := cmdFlags.Args()
	if len(processNames) == 0 {
		c.Ui.Error("At least one process name must be supplied.")
		c.Ui.Error("")
		c.Ui.Error(c.Help())
		return 1
	}

	client, err := RPCClient(*rpcAddr)
	if err != nil {
		c.Ui.Error("Error connecting to Watchdog agent")
		return 1
	}
	defer client.Close()

	pids, err := client.Start(processNames...)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error starting processes: %s", err))
		return 1
	}

	c.Ui.Output(fmt.Sprintf(
		"Successfully started processes with PIDs: %v", pids))

	return 0
}

func (c *StartCommand) Synopsis() string {
	return "Start a process"
}
