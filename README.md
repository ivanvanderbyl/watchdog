# Watchdog

Watchdog is a network connected process manager for systems where you don't want to
ever have to think about what is running and where its output is being sent.

Watchdog supports multiple output drains to send your process's logs to `l2met`, `logentries`, `file` and any other http backed logging interface.

Watchdog is designed to run as a daemon and act as the parent process to the processes it manages. Watchdog is written in Go, and dispatches a separate goroutine for each process under management, using channels to route the output to the necessary drains.

Watchdog consists of a single, easy to deploy binary that acts as both the agent and control interface.

One of the main design goals of Watchdog is to be easy to configure, as such we use `toml`, a modern replacement for the classic config file providing a clean, concise, and extensive set of constructs for describing process bahaviour.

Internally Watchdog uses an RPC interface for communicating with the agent process which can easily be programmed against.
