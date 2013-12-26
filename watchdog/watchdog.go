package watchdog

// The purpose of this package is to act as a registry for processes
// and proxy commands to processes and respond to events from processes, even
// if the event handling is only logging events.
//
// It is also responsible for sending log output to drain channels.
//
// All exported methods in this package are designed to be interacted with by the `agent` package.
//
// In typical operation this package would only be run once as a daemon per host.
