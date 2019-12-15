# RemoteRotator

![Build Status](https://github.com/dh1tw/remoteRotator/workflows/Cross%20Platform%20build/badge.svg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/dh1tw/remoteRotator)](https://goreportcard.com/report/github.com/dh1tw/remoteRotator)
[![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://img.shields.io/badge/license-MIT-blue.svg)
[![Coverage Status](https://coveralls.io/repos/github/dh1tw/remoteRotator/badge.svg?branch=master)](https://coveralls.io/github/dh1tw/remoteRotator?branch=master)
[![Downloads](https://img.shields.io/github/downloads/dh1tw/remoteRotator/total.svg?maxAge=1800)](https://github.com/dh1tw/remoteRotator/releases)

![Alt text](https://i.imgur.com/lcHhslZ.png "remoteRotator WebUI")

remoteRotator is a command line application which makes your azimuth / elevation
antenna rotators available on the network. It available for Linux/Windows/MacOS and written in the programing language [Go](https://golang.org).

**ADVICE**: This project is **under development**. The parameters and the ICD
are still **not stable** and subject to change until the first major version
has been reached.

## Supported Rotators

- [Yaesu GS232A](http://www.yaesu.com/indexVS.cfm?cmd=DisplayProducts&ProdCatID=104&encProdID=79A89CEC477AA3B819EE02831F3FD5B8)
- [EA4TX ARS (implements Yaesu GS232A)](http://ea4tx.com/en/)
- Dummy rotator

## Supported Transportation Protocols

- [NATS](https://nats.io) (preferred)
- HTTP + Websockets + MDNS

## License

remoteRotator is published under the permissive [MIT license](https://github.com/dh1tw/remoteRotator/blob/master/LICENSE).

## Download

You can download a tarball / zip archive with the compiled binary for

- MacOS (amd64)
- Linux (arm,386,amd64)
- Windows (386,amd64)

 from the [releases](https://github.com/dh1tw/remoteRotator/releases) page.

remoteRotator works well on SoC boards like the Raspberry / Orange / Banana Pis.
The application is a self contained single executable.

## Dependencies

remoteRotator does not have any runtime dependencies.

## Getting started

remoteRotator provides a series of nested commands and flags.

```bash
$ ./remoteRotator
```

```text
Network interface for Rotators

Usage:
  remoteRotator [command]

Available Commands:
  enumerate   discover and list all available rotators on the network
  help        Help about any command
  server      remoteRotator Server
  version     Print the version number of remoteRotator
  web         webserver providing access to all rotators on the network

Flags:
      --config string   config file (default is $HOME/.remoteRotator.yaml)
  -h, --help            help for remoteRotator

Use "remoteRotator [command] --help" for more information about a command.
```

So let's fire up a remoteRotator server for your rotator:

First, identify the serial port to which your rotator is connected. On Windows
this will be something like `COMx` (e.g. `COM3`), on Linux / MacOS it will be
a device in the `/dev/` folder (e.g. `/dev/ttyACM0`).

All parameters can be set either in a config file (see below) or through pflags.
To get a list of supported flags for the lan server, execute:

```bash
$ ./remoteRotator server lan --help
```

```text
The lan server allows you to expose a rotator on your local area network.
By default, the rotator will only be listening on the loopback adapter. In
order to make it available and discoverable on the local network, a network
connected adapter has to be selected.

remoteRotator supports access via TCP, emulating the Yaesu GS232 protocol
(disabled by default) and through a web interface (HTTP + Websocket).

You can select the following rotator types:
1. Yaesu (GS232 compatible)
2. Dummy (great for testing)

remoteRotator allows to assign a series of meta data to a rotator:
1. Name
2. Azimuth/Elevation minimum value
3. Azimuth/Elevation maximum value
4. Azimuth Mechanical stop

These metadata enhance the rotators view (e.g. showing overlap) in the web
interface and can also help to limit for example the rotators range if it does
not support full 360°.

Usage:
  remoteRotator server lan [flags]

Flags:
      --azimuth-max int        metadata: maximum azimuth (in deg) (default 360)
      --azimuth-min int        metadata: minimum azimuth (in deg)
      --azimuth-stop int       metadata: mechanical azimuth stop (in deg)
  -b, --baudrate int           baudrate (default 9600)
      --discovery-enabled      make rotator discoverable on the network (default true)
      --elevation-max int      metadata: maximum elevation (in deg) (default 180)
      --elevation-min int      metadata: minimum elevation (in deg)
      --has-azimuth            rotator supports Azimuth (default true)
      --has-elevation          rotator supports Elevation
  -h, --help                   help for lan
      --http-enabled           enable HTTP Server (default true)
  -w, --http-host string       Host (use '0.0.0.0' to listen on all network adapters) (default "127.0.0.1")
  -k, --http-port int          Port for the HTTP access to the rotator (default 7070)
  -n, --name string            Name tag for the rotator (default "myRotator")
      --pollingrate duration   rotator polling rate (default 1s)
  -P, --portname string        portname / path to the rotator (e.g. COM1) (default "/dev/ttyACM0")
      --tcp-enabled            enable TCP Server
  -u, --tcp-host string        Host (use '0.0.0.0' to listen on all network adapters) (default "127.0.0.1")
  -p, --tcp-port int           TCP Port (default 7373)
  -t, --type string            Rotator type (supported: yaesu, dummy (default "yaesu")

Global Flags:
      --config string   config file (default is $HOME/.remoteRotator.yaml)
```

So in order to launch remoteRotator on Windows with a Yaesu rotator connected at
COM3 an having the web HTTP server listening on your local network, we would call:

```bash
$ remoteRotator.exe server lan -w "0.0.0.0" -P "COM3" -t yaesu
```

```text
no config file found
2017/12/08 16:50:25 added rotator (myRotator)
2017/12/08 16:50:25 Listening on 0.0.0.0:7070 for HTTP connections
```

## Connecting via TCP / Telnet

If you have an application (e.g. [arsvcom](https://ea4tx.com/en/arsvcom/) or
[pstrotator](http://www.qsl.net/yo3dmu/index_Page346.htm)) which can talk to
a Yaesu compatible rotator, you can point that application to remoteRotator's
builtin TCP server (although disabled by default).

Let's start a dummy rotator instance on Linux and enable the build-in TCP
server:

``` bash
$ ./remoteRotator server lan -t dummy --tcp-enabled
```

``` text
no config file found
2017/12/08 16:50:25 added rotator (myRotator)
2017/12/08 16:50:25 listening on 127.0.0.1:7070 for HTTP connections
2017/12/08 16:50:25 listening on 127.0.0.1:7373 for TCP connections
2017/12/08 16:50:25 discovery disabled; the HTTP server must listen on an accessible network interface (e.g. 0.0.0.0)
```

For testing, we connect directly via telnet:

``` text
$ telnet localhost 7373
Trying ::1...
Connected to localhost.
Escape character is '^]'.

?>C
+0303
C2
+0303+0000
M310
+0303+0000
+0304+0000
+0305+0000
+0306+0000
+0307+0000
+0307+0000
+0308+0000
+0309+0000
+0310+0000
```

## Web Interface

![Alt text](https://i.imgur.com/wPup7BJ.png "remoteRotator WebUI")

A more comfortable way of accessing the rotator is through a web Interface.
You can specify the host and port in the settings above, or deactivate the
built-in webserver if you don't need it.

The red arrow indicates the heading of the rotator and the yellow arrow
indicates the preset value to which the rotator will turn to. The yellow arrow
disappears when the desired direction has been reached.

The dotted red line indicates the mechanical stop of the rotator.
A green arc segment indicates a limited turning radius for this rotator.
A blue arc segment indicates the mechanical overlap supported by this rotator.
These indicators are just visual helpers and are configurable though command line
flags or in the config file.

## Web Interface (Aggregator)

![Alt text](https://i.imgur.com/lcHhslZ.png "remoteRotator WebUI")

If you have multiple rotators, you might want to use the dedicated aggregation
web server. The following example starts the webserver on port 6005 and listens
on all network interfaces.

``` text
$ remoteRotator web -w "0.0.0.0" -k 6005
```

The Webserver automatically discovers the available remoteRotator instances
in your local network and adds them from the web interface. Depending on which transport you have chosen, the discovery process is either done through MDNS (lan)
or NATS. The discovery functionality doesn't require any configuration.

## Config file

The repository contains an example configuration file. By convention it is called
`.remoteRotator.[yaml|toml|json]` and is located by default either in the
home directory or the directory where the remoteRotator executable is located.
The format of the file can either be in
[yaml](https://en.wikipedia.org/wiki/YAML),
[toml](https://github.com/toml-lang/toml), or
[json](https://en.wikipedia.org/wiki/JSON).

The first line after starting remoteRotator will indicate if / which config
file has been found.

If you have several rotators, you have to create a configuration file for
each of them and specify them with the --config flag.

Priority:

1. Pflags (e.g. -p 4040 -t dummy)
2. Values from config file
3. Default values

## Behaviour on Errors

If an error occurs from which remoteRotator can not recover, the application
exits. This typically happens when the connection with the rotator has been
lost or if the rotator is not responding anymore.
It is recommended to execute remoteRotator as a service under the supervision
of a scheduler like [systemd](https://en.wikipedia.org/wiki/Systemd).

## Bug reports, Questions & Pull Requests

Please use the Github [Issue tracker](https://github.com/dh1tw/remoteRotator/issues)
to report bugs and ask questions! If you would like to contribute to remoteRotator,
[pull requests](https://help.github.com/articles/creating-a-pull-request/) are
welcome! However please consider to provide unit tests with the PR to verify
the proper behavior.

If you file a bug report, please include always the version of remoteRotator
you are running:

``` bash
$ remoteRotator.exe version
```

``` text
copyright Tobias Wellnitz, DH1TW, 2017
remoteRotator Version: 0.1.0, darwin/amd64, BuildDate: 2017-09-04T00:58:00+02:00, Commit: 338ff13
```

## Documentation

The auto generated documentation can be found at
[godoc.org](https://godoc.org/github.com/dh1tw/remoteRotator).

## How to build

In order to compile remoteRotator from the sources, you need to have
[Go >= 1.13](https://golang.org) installed and configured.

This his how to checkout and compile remoteRotator under Linux/MacOS:

``` bash
$ go get -d github.com/dh1tw/remoteRotator
$ cd $GOPATH/src/github.com/dh1tw/remoteRotator
$ make install-deps
$ make generate
$ make
```

## How to execute the tests

All critical packages have their own set of unit tests. The tests can be
executed with the following commands:

``` bash
$ cd $GOPATH/src/github.com/dh1tw/remoteRotator
$ go test -v -race ./...
```

The data race detector might not be available on all platforms / operating
systems.
