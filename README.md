[![Go Report Card](https://goreportcard.com/badge/github.com/blacknon/lsshell)](https://goreportcard.com/report/github.com/blacknon/lsshell)

lsshell
===

`lsshell` is a TUI (Text-based User Interface) tool for managing parallel SSH sessions with an easy-to-use list selection interface.
It allows you to execute commands across multiple remote servers simultaneously, making it ideal for system administrators and developers managing multiple machines.

This tool is a related project of [lssh](https://github.com/blacknon/lssh). It uses the same configuration file as **lssh**.

## Features

- **Parallel SSH Execution**: Execute commands on multiple remote hosts simultaneously.
- **TUI Interface**: A user-friendly text-based interface for selecting hosts and managing sessions.
- **Flexible Configuration**: Easily configurable for different environments.

## Installation

To install `lsshell`, clone the repository and build it using Go.

```bash
git clone https://github.com/blacknon/lsshell.git
cd lsshell
go build
```

## Usage

To start lsshell, run the following command.

```bash
lsshell
```

You can then select hosts and execute commands across multiple sessions.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
