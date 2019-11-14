package cmd

import (
	"fmt"
    "io/ioutil"
    "gopkg.in/yaml.v2"
)

//Command -
type Command struct {
    Name string
    Path string
}
//LogFile -
type LogFile struct {
    Name string //Must be unique within a host running this app. Used to save the tail pos
    Paths []string
    Timelayout string //Parse the match below into go time object
    Timepattern string //extract the timestamp part into a timeStr which is fed into the Timelayout
	Timeadjust string //If the time extracted string miss some info (like year or zone etc) this string will be appended to the string
    Pattern string //will be matched to extract the HOSTNAME APP-NAME MSG part of the line.
    Multilineptn string //detect if the line is part of the previous line
    Appname string //Overrite the appname of the logfile if not empty
}
//AppConfig -
type AppConfig struct { //Why do I have to tag every field! Because yaml driver automatically lowercase the field name to look into the yaml file <yuk>
    Port int
    Commands []Command
    JwtKey string
    Logfiles []LogFile
    Serverurl string
    Logdbpath string
}

//Config - Global
var Config AppConfig

//TimeISO8601LayOut
const (
    TimeISO8601LayOut = "2006-01-02T15:04:05-0700"
)

//GenerateDefaultConfig -
func GenerateDefaultConfig(opt ...interface{}) (e error) {
    defaultConfig := `
port: 8000
# Used in client mode to send to the server
serverurl: http://localhost:8000
jwtkey: ChangeThisKeyInYourSystem
logdbpath: logs.db
# commands list to allow remote execution.
commands:
    - name: example_ls
      path: /bin/ls
logfiles:
    - name: syslog
      paths:
        - /var/log/syslog
      timelayout: "Jan 02 15:04:05 2006 MST"
      timepattern: '^([a-zA-Z]{3,3} [\d]{0,2} [\d]{2,2}\:[\d]{2,2}\:[\d]{2,2}) '
      timeadjust: "2019 AEST"
      pattern: '([^\s]+) ([^\s]+) (.*)$'
      multilineptn: '([^\s]+.*)$'
      appname: ""
`
tailSimpleConfig := `
port: 8000
# Used in client mode to send to the server
serverurl: %s
jwtkey: %s
logdbpath: logs.db
# commands list to allow remote execution.
commands:
    - name: example_ls
      path: /bin/ls
logfiles:
    - name: SimpleFileParser
      paths:
        - %s
      timelayout: "Jan 02 15:04:05 2006 MST"
      timepattern: ''
      timeadjust: ""
      pattern: '^([^\s]+.*)$'
      multilineptn: '^[\s]+([^\s]+.*)$'
      appname: '%s'
`
    var fPath, configContent, serverurl, jwtkey, logfile, appname string
    configContent = defaultConfig

    for i, v := range(opt) {
        if i % 2 == 0 {
            key := v.(string)
            switch key {
            case "file":
                fPath = opt[i+1].(string)
            case "serverurl":
                serverurl = opt[i+1].(string)
            case "jwtkey":
                jwtkey = opt[i+1].(string)
            case "logfile":
                logfile = opt[i+1].(string)
            case "appname":
                appname = opt[i+1].(string)
            }
        }
    }
    if logfile != "" {
        configContent = fmt.Sprintf(tailSimpleConfig, serverurl, jwtkey, logfile, appname)
    } else {
        configContent = defaultConfig
    }
    err := ioutil.WriteFile(fPath, []byte(configContent), 0600)
    if err != nil {return err}
    return LoadConfig(fPath)
}

//LoadConfig -
func LoadConfig(fPath string) (e error) {
    yamlStr, e := ioutil.ReadFile(fPath)
    if e != nil {
        return e
    }
    e = yaml.Unmarshal(yamlStr, &Config)
    return e
}