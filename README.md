# nsre
Nagios Simple Remote Execution and Log Server

# Why

I need a fast and lightweight log server to store log in general *AND* a simple
remote execution point (similar to nsca or nrpe) to answer nagios check.

nsca and nsre does not meet as they dont store logs. And they are a bit big with
many other features I dont need.

For Logging I dont want elasticsearch because I have no budget for big server.
And it does not has the remote executor ...

So my requirements are pretty unique but suitable in a DevOps case that a place
to sotre pretty much all kind of logs and offer a simple search interface. And a
remote executor for nagios plugin. And can run a AWS t2.nano instance.

# Quick Start

## Build

As of this time I dont use fts yet so you can remove it. Maybe in the future?

```
go build --tags "icu json1 fts5 secure_delete" -ldflags='-s -w'
```

On windows you need to install and setup golang, git and
[msys2](http://www.msys2.org/). Golang sqlite3 driver uses cgo thus you need
gcc, install from msys2. Please refer to each documentation guide to get it done.

Once the build system is installed, use the same go build above to build teh
software.

## Usage

The app is controled by option `-m` modename. Run with `-h` for more details.

### Create sample config file

`nsre -c nsre.yaml -m setup `

will create a sample nsre.yaml config file. Then edit the file to suit your
needs.

### Server mode

For executing remote command and get logs.

`nsre -c nsre.yaml -m server`

If you want to also harvest log in the same server running this then use

`nsre -c nsre.yaml -m tailserver`

### Log shipping in server mode

This will tail the log and ship to server.

`nsre -c nsre.yaml -m tail -tailf`

Remove `-tailf` so it will tail the log and exit.

### Add hoc log shipping

For logs does not have timestamp - etc.. It runs once and exit.

First create config. Then run it. Below example is what I use in my jenkins file
to ship the build log.


```
nsre -m setup -c /tmp/nsre-\$\$.yaml -url ${nsre_url} -f ${BUILD_TAG}.log, -jwtkey ${NSRE_JWT_API_KEY} -appname ${BUILD_TAG}
nsre -m tail -c /tmp/nsre-\$\$.yaml
rm -f /tmp/nsre-\$\$.yaml
```

Please look at the source code for how it parses logs.

### Remote Execution for Nagios

First define the command name in the config file and command path in the server
part (where the actual command is run).

At the client (the nagios server which runs an active check) run

```
# This will run and return a text output, first line is exit code, and the rest
# the stdout of the remote command. Used for debugging
nsre -c config_file -m client -cmd <command_name>
# This will return what nagios expects it to return - a exit code to show the
# error or OK and output as it is. Use as nagios plugin.
nsre -c config_file -m nagios -cmd <command_name>
```

It does not take any options. You need to define all options in each command
name at the server and control it output.


### Search Log Web GUI

The url would be (depending on if you configure ssl or not)
`http(s)://hostname:port/searchlog`

### Confusion between client/server mixup

I do as it has some mixture between usages and it can act both at the same time.
However look into the cmd/config.go for more details.

In short in the yaml file

- `Port` is server part, which the program in server mode will listen.
- `Serverurl` is client part, which the program in client mode will use to post
  request to (store log or run command)

So in `-m tailserver`  mode you may send the log to some other real log server
and it listens for the remote command to execute on `Port` if the `Serverurl` is
not itself. This is intended usages.


END OF DOCUMENT
