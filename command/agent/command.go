package agent

import (
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// gracefulTimeout controls how long we wait before forcefully terminating
var gracefulTimeout = 3 * time.Second

// Command is a Command implementation that runs a Watchdog agent.
// The command will not end unless a shutdown message is sent on the
// ShutdownCh. If two messages are sent on the ShutdownCh it will forcibly
// exit.
type Command struct {
	Ui         cli.Ui
	ShutdownCh <-chan struct{}
	args       []string
}

func (c *Command) Run(args []string) int {
	c.Ui = &cli.PrefixedUi{
		OutputPrefix: "---> ",
		InfoPrefix:   "     ",
		ErrorPrefix:  " ! > ",
		Ui:           c.Ui,
	}

	c.Ui.Output("Starting Watchdog...")

	// Parse our configs
	c.args = args

	// Setup watchdog
	agent := NewAgent()
	if agent == nil {
		return 1
	}
	defer agent.Shutdown()

	// Start the agent after the handler is registered
	if err := agent.Start(); err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to start the Watchdog agent: %v", err))
		return 1
	}

	c.Ui.Info("Watchdog agent running!")

	// Wait for exit
	return c.handleSignals(agent)
}

// readConfig is responsible for setup of our configuration using
// the command line and any file configs
func (c *Command) readConfig() *Config {
	var cmdConfig Config
	var configFiles []string
	cmdFlags := flag.NewFlagSet("agent", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }

	cmdFlags.Var((*AppendSliceValue)(&configFiles), "config-file",
		"json file to read config from")
	cmdFlags.Var((*AppendSliceValue)(&configFiles), "config-dir",
		"directory of json files to read")
	cmdFlags.StringVar(&cmdConfig.LogLevel, "log-level", "", "log level")
	cmdFlags.StringVar(&cmdConfig.RPCAddr, "rpc-addr", "",
		"address to bind RPC listener to")

	if err := cmdFlags.Parse(c.args); err != nil {
		return nil
	}

	config := DefaultConfig
	if len(configFiles) > 0 {
		fileConfig, err := ReadConfigPaths(configFiles)
		if err != nil {
			c.Ui.Error(err.Error())
			return nil
		}

		config = MergeConfig(config, fileConfig)
	}

	config = MergeConfig(config, &cmdConfig)
	return config
}

// handleSignals blocks until we get an exit-causing signal
func (c *Command) handleSignals(agent *Agent) int {
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	// Wait for a signal
	var sig os.Signal
	select {
	case s := <-signalCh:
		sig = s
	case <-c.ShutdownCh:
		sig = os.Interrupt
	case <-agent.ShutdownCh():
		// Agent is already shutdown!
		return 0
	}
	c.Ui.Output(fmt.Sprintf("Caught signal: %v", sig))

	// Check if we should do a graceful leave
	graceful := false
	if sig == os.Interrupt {
		graceful = true
	} else if sig == syscall.SIGTERM {
		graceful = true
	}

	// Bail fast if not doing a graceful leave
	if !graceful {
		return 1
	}

	// Attempt a graceful leave
	gracefulCh := make(chan struct{})
	c.Ui.Output("Gracefully shutting down agent...")
	go func() {
		if err := agent.Shutdown(); err != nil {
			c.Ui.Error(fmt.Sprintf("Error: %s", err))
			return
		}
		close(gracefulCh)
	}()

	// Wait for leave or another signal
	select {
	case <-signalCh:
		return 1
	case <-time.After(gracefulTimeout):
		return 1
	case <-gracefulCh:
		return 0
	}
}

func (c *Command) Synopsis() string {
	return "Runs the Watchdog agent"
}

func (c *Command) Help() string {
	helpText := `
Usage: watchdog agent [options]

  Starts the Watchdog agent and runs until an interrupt is received. The agent
  will run in the foreground as it is designed to run under the supervision of
  the OS process manager Upstart/launchd.

Options:

  -config-file=foo         Path to a JSON or TOML file to read configuration from.
                           This can be specified multiple times.
  -config-dir=foo          Path to a directory to read configuration files
                           from. This will read every file ending
                           in ".json" or ".toml" as configuration in this
                           directory in alphabetical order.
  -log-level=info          Log level of the agent (debug,info,warn,error).
  -rpc-addr=127.0.0.1:6673 Address to bind the RPC listener.

`
	return strings.TrimSpace(helpText)
}
