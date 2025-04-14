# Onkyo Controller

Highly optimized onkyo eiscp protocol implementation for TX-L20D model with minimal set of features.

![showcase](./assets/showcase.gif)

## Features
- Power on/off control
- Volume and bass adjustment via digital crown
- Profile switching (audio source, volume settings, bass presets)

## Implementation
Go-based server implementing the onkyo-eiscp protocol with:
- CLI interface for scripts and automation
- JSON API for the Apple Watch app and Apple Shortcuts

## Usage
```
> go build -o target/onkyo cmd/cli/main.go cmd/cli/chat.go
> cp target/onkyo ~/.local/bin/
> export ONKYO_HOST="10.205.0.163"

> onkyo help
NAME:
   onkyo - Onkyo TX-L20D client

USAGE:
   onkyo [global options] [command [command options]]

COMMANDS:
   power      Control device power
   volume     Control volume settings
   subwoofer  Control subwoofer settings
   source     Control input source
   chat       Chat with onkyo using raw eiscp messages
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --host value, -H value  Onkyo host ip address (default: "127.0.0.1") [$ONKYO_HOST]
   --port value, -P value  Onkyo host port (default: "60128") [$ONKYO_PORT]
   --help, -h              show help

> onkyo chat
Chat session with Onkyo TX-L20D established.
Type EISCP commands or 'exit' to quit.
Use Ctrl+C or Ctrl+D to terminate the session.
Use arrow up/down to navigate command history.

> SWL+04
TX-L20D: SWL+04

> SWLDOWN
TX-L20D: SWL+03

> ^D
Terminating chat session...

> onkyo power off
```

## Acknowledgments
Based on amazing work from [onkyo-eiscp](https://github.com/miracle2k/onkyo-eiscp)
