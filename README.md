# Watchdog

Watchdog is a modern rethinking of a process manager for running applications and services at scale with minimal condiguration and deployment effort.

### Architecture

Watchdog is delivered as a single binary which acts as both the control interface and background agent. It is designed to run as a foreground process under Upstart or Launchd (See the examples directory for startup config examples).

To start the background agent, run:

```sh
watchdog agent
```

This will also start the RPC interface on `127.0.0.1:6673`.

### Configuration

Watchdog is designed to be easy to configure either by hand or by machine. As such it supports two interchangable configuration formats; `JSON` and [TOML](https://github.com/mojombo/toml). Configuration files can be located anywhere and registred at runtime, at which point they will be watched for changes by Watchdog and changes automatically applied on save.

To register a new process configuration:

```sh
watchdog register /path/to/myprocess.json
```

If your process is configured `start_on_load` it will be started immediately, otherwise it will be registered and not started until you manually start the process.

### Process Control

Watchdog CLI includes the usual methods for starting, stopping and restarting processes. The real power here comes from the process configuration which allows you to configure how processes are signalled to exit and how long to wait between restarting to avoid overloading the system.

Start a process:

```sh
watchdog start myprocess
```

Stop a process:

```sh
watchdog stop myprocess
```

Restart a process:

```sh
watchdog restart myprocess
```

### Tailing process logs

It is expected that any useful process output will be written to `stdout` or `stderr` as per the usual [12 Factor App](http://12factor.net/logs) setup.

You can configure custom log drains to have process output directed to external services like `Librato`, `l2met`, `LogEntries`, `Loggly`, `file`.

You can also use the CLI to tail process logs in realtime:

```sh
watchdog logs -tail myprocess
```

### Output drains

Watchdog supports multiple output drains on a per process basis, allowing to you effortlessly ship output to any of the following services:

- Librato
- l2Met
- LogEntries
- Loggly

## Design Goals

*Some of the above mentioned features are still being developed*

Watchdog is primarily designed to solve the problem of configuration drift across platforms for managing similar process types which often occures when working with different operating systems. At App.io we run OSX and Ubuntu. Our entire stack is configured using a configuration management tool and rarely ever do we connect directly to a machine to configure or probe anything.

As such Watchdog is designed to output both machine readable (JSON) and human readable data, making automated configuration easy.

## API Documentation

[API Documentation](http://godoc.org/github.com/appio/watchdog)

## License

Watchdog is licensed under the [Mozilla Public License, version 2.0](LICENSE)
