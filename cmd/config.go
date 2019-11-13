package cmd

import (
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
    Path string
    Timelayout string //Parse the match below into go time object
    Timepattern string //extract the timestamp part into a timeStr which is fed into the Timelayout
	Timeadjust string //If the time extracted string miss some info (like year or zone etc) this string will be appended to the string
    Pattern string //will be matched to extract the HOSTNAME APP-NAME MSG part of the line.
    Multilineptn string //detect if the line is part of the previous line
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
func GenerateDefaultConfig(fPath string) (e error) {
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
      path: /var/log/syslog
      timelayout: "Jan 02 15:04:05 2006 MST"
      timepattern: '^([a-zA-Z]{3,3} [\d]{0,2} [\d]{2,2}\:[\d]{2,2}\:[\d]{2,2}) '
      timeadjust: "2019 AEST"
      pattern: '([^\s]+) ([^\s]+) (.*)$'
      multilineptn: '^[\s\t]+([^\s]+.*)$'
`
    err := ioutil.WriteFile(fPath, []byte(defaultConfig), 0600)
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