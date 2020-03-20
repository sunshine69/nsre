# nsre
Nagios Simple Remote Execution and Log Server

# Why

I need a fast and lightweight log server to store log in general *AND* a simple
remote execution point (similar to nsca or nrpe) to answer nagios check.

nsca and nrpe does not meet as they dont store logs. And they are a bit big with
many other features I don't need.

For Logging I don't want elasticsearch because I have no budget for big server.
And it does not has the remote executor ...

So my requirements are pretty unique but suitable in a DevOps case that a place
to sotre pretty much all kind of logs and offer a simple search interface. And a
remote executor for nagios plugin. And can run a AWS t2.nano instance.

This can be a foundation for many sysadmin utilities due to the power of golang!

# Quick Start

## Build

As of this time I dont use fts yet so you can remove it. Maybe in the future?

```
go build --tags "icu json1 fts5 secure_delete" -ldflags='-s -w'
```

To build a static binary on Linux using alpine image just do (you may need to edit the build-static.sh to suite your need)

```
docker build -t golang-alpine-build:latest -f Dockerfile.nsre-alpine-build  .
./build-static.sh
```


On windows you need to install and setup golang, git and
[msys2](http://www.msys2.org/). Golang sqlite3 driver uses cgo thus you need
gcc, install from msys2. Please refer to each documentation guide to get it done.

Once the build system is installed, use the same go build above to build the
software.

## Usage

The app is controled by option `-m` modename. Run with `-h` for more details.

### Create sample config file

`nsre -c nsre.yaml -m setup `

will create a sample nsre.yaml config file. Then edit the file to suit your
needs.

The application uses OAuth to authenticate user in the websearch interface.
The provider supported is google.com. In the config you would see the section
below:


```
# Your google api client id and secret.
appgoogleclientid: ""
appgoogleclientsecret: ""

# Session to encrypt session cookie. Be sure it is secret and having enough
# length.

sessionkey: ""

# If your company has custom domain when using google, eg. `name@somecompany.com`
# and you want to authorise all people to login add `somecompany.co` to
# `authorizeddomain`

authorizeddomain:
  somecompany.com: true

# In addition to the above, if you want to authorize by person you need to
# manually insert that person in the table `user`. At the server run the command
# `sqlite3 <database_name, eg. logs.db>`
# .schema will show table user.
# Insert a record in, then that user with matching email address will be
# allowed.

# The system first check the user table first and then check the authorized
# domain. If no user in the table, it will check the domain.


```

### Server mode

For executing remote command and get logs.

`nsre -c nsre.yaml -m server`

If you want to also harvest log in the same server running this then use

`nsre -c nsre.yaml -m tailserver -tailf`

This also enabled awslogs fetching if you set any awslogs entries in config. See
below.

### Log shipping in server mode

This will tail the log and ship to server.

`nsre -c nsre.yaml -m tail -tailf`

Remove `-tailf` so it will tail the log and exit.

### AWSLogs support

In the config you can see one entry which has empty list awslogs: []

This is for you to define several aws cloudwatch log entries to query. Sample
below

```
awslogs:
    - loggroupname: '/aws/ecs/uat'
      streamprefix:
        - 'xvt-report-pdf'
        - 'errcd-wa'
        - 'api'
      filterptn: ''
      profile: 'errcd_wa'
      region: 'ap-southeast-2'
      period: '5m'
    - loggroupname: '/aws/ecs/int'
      streamprefix:
        - 'api'
        - 'errcd-wa'
        - 'xvt-report-pdf'
      filterptn: ''
      profile: 'errcd_wa'
      region:
      period: '5m'

```

It is self explanatory! The `period` specify how long we sleep and wake up to
fetch logs, and how long in the past since the time you start the server to
fetch log.


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

### Running as service

On linux I use systemd and a wrapper script looks like below.

```
#!/bin/bash

export SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd $SCRIPT_DIR

export HOME=/root
nohup ./nsre $* &

```

Systemd unit at

```
cat /etc/systemd/system/nsre.service
[Unit]
Description=NSRE Server
Documentation=https://github.com/sunshine69/nsre
#After=network-online.target firewalld.service
#Wants=network-online.target

[Service]
Type=forking

StandardOutput=file:/var/nsre/nsre.log
StandardError=file:/var/nsre/nsre-error.log
#SyslogIdentifier=<your program identifier> # without any quote

WorkingDirectory=/var/nsre
#PIDFile=/run/nagios-api.pid
ExecStart=/var/nsre/nsre.sh -c nsre.yaml -m tailserver -tailf
#ExecReload=/bin/kill -s HUP $MAINPID
# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNOFILE=infinity
LimitNPROC=infinity
LimitCORE=infinity
# Uncomment TasksMax if your systemd version supports it.
# Only systemd 226 and above support this version.
#TasksMax=infinity
TimeoutStartSec=0
# kill only the nagios-api process, not all processes in the cgroup
KillMode=process
# restart the nagios-api process if it exits prematurely
Restart=on-failure
StartLimitBurst=3
StartLimitInterval=60s

[Install]
WantedBy=multi-user.target
```

On windows I use [nssm](https://nssm.cc/usage)

```
nssm install nsre "C:\ansible_install\nsre\nsre.exe" -c "C:\ansible_install\nsre\nsre.yaml" -m tail -tailf
nssm set nsre AppDirectory "C:\ansible_install\nsre"
```

### Sample complete config file

To configure the time pattern to parse the log line first look at this struct comment to understand these values (copy from config.go)

```
type LogFile struct {
    Name string //Must be unique within a host running this app. Used to save the tail pos
    Paths []string
    Timelayout string //Parse the match below into go time object
    Timepattern string //extract the timestamp part into a timeStr which is fed into the Timelayout
    Timeadjust string //If the time extracted string miss some info (like year or zone etc) this string will be appended to the string. It may have special string 'syslog' to auto adjust for syslog time stamp. If it contains a golang timelayout token with one extra space at the end of the string (eg. '2004 ') then these token will be parsed as the current for example year.
    Timestrreplace []string //Do search/replace the capture before parse time. As go does not support , aas sec fraction this is to work around for this case.
    Pattern string //will be matched to extract the HOSTNAME APP-NAME MSG part of the line.
    Multilineptn string //detect if the line is part of the previous line
    Excludepatterns []string //If log line match this pattern it will be excluded
    Includepatterns []string
    Appname string //Overrite the appname of the logfile if not empty
}
```

The `logfiles` in the complete nsre.yaml config file below is followed the above struct.

```
# Server listen port
port: 8000

# Commands definitions list
commands:
- name: example_ls
  path: /bin/ls
- name: ping
  path: /bin/echo pong

# JWT key for the client server communications. Need to be the same for server
# and all client (log shipping and nagios check command)

jwtkey: <YOUR JWT KEY>

logfiles:
- name: errcd-iis-activity-tas
  paths:
# Full path or glob pattern also works
  - D:\nsw-errcdconsolidation-tas-test\logs\DAPIS\activity.log
  - D:\ERRCD\nsw_errcdconsolidation_aus-qa-rtr\*\tools\RTR.DataProcessing\*.log
  timelayout: 2006-01-02 15:04:05.999 MST
  timepattern: '^([\d]{4,4}\-[\d]{2,2}\-[\d]{2,2} [\d]{2,2}\:[\d]{2,2}\:[\d]{2,2}[,]{1,1}[\d]{3,3})[\s]+'
  timestrreplace: [",", "."]
  timeadjust: UTC
  pattern: ([^\s]+.*)$
  multilineptn: ([^\s]*.*)$
  appname: "errcd-activity"

# This is for client mode
serverurl: <Your server url>
ignorecertificatecheck: false

# Database config
logdbpath: logs.db
dbtimeout: 45s

# If provided the server will enable https on the port specified above
sslcert: ""
sslkey: ""

# At server and client before saving log record the data will be search/replace
# using this regex pattern to filter sensitive password getting in.

# The non capture group will be replaced by a string DATA_FILTERED.

passwordfilterpattern:
  - ([Pp]assword|[Pp]assphrase)['"]*[\:\=]*[\s\n]*[^\s]+[\s;]
  - <more pattern>

# See example above
awslogs: []
```

### Twilio call and sms proxying

See comment in `cmd/twilio.go` for more information.

### AWS SNS subscription endpoint to send sms/voice message

See the comment in `cmd/aws-sns.go`.

### External program to ship log

If you do nto sue nsre in client mode to tail and ship log, but want to send log directly from your app then read on.

#### Authentication
Need to make a http-header 'Token' to contains the JWT Token. The signature needs to be verified to be the same shared key configured in the server.

### Data
Body of request should be a json data matching the json schema below.

```
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "$ref": "#/definitions/LogData",
  "definitions": {
    "LogData": {
      "required": [
        "Timestamp",
        "Datelog",
        "Host",
        "Application",
        "Logfile",
        "Message"
      ],
      "properties": {
        "Timestamp": {
          "type": "integer"
        },
        "Datelog": {
          "type": "integer"
        },
        "Host": {
          "type": "string"
        },
        "Application": {
          "type": "string"
        },
        "Logfile": {
          "type": "string"
        },
        "Message": {
          "type": "string"
        }
      },
      "additionalProperties": false,
      "type": "object"
    }
  }
}
```

END OF DOCUMENT
