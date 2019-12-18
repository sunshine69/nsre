package cmd

import (
	"regexp"
	"strconv"
	"fmt"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2"
	"time"
	"github.com/bvinc/go-sqlite-lite/sqlite3"
	"os"
	"strings"
	"log"
    "io/ioutil"
    "gopkg.in/yaml.v2"
    "github.com/gorilla/sessions"
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
    Timeadjust string //If the time extracted string miss some info (like year or zone etc) this string will be appended to the string. It may have special string 'syslog' to auto adjust for syslog time stamp
    Timestrreplace []string //Do search/replace the capture before parse time. As go does not support , aas sec fraction this is to work around for this case.
    Pattern string //will be matched to extract the HOSTNAME APP-NAME MSG part of the line.
    Multilineptn string //detect if the line is part of the previous line
    Excludepatterns []string //If log line match this pattern it will be excluded
    Includepatterns []string
    Appname string //Overrite the appname of the logfile if not empty
}
//AppConfig -
type AppConfig struct { //Why do I have to tag every field! Because yaml driver automatically lowercase the field name to look into the yaml file <yuk>
    Serverdomain string
    Port int //Server mode - Port to listen for web gui, log harvesting and Command executor
    Commands []Command
    JwtKey string
    Logfiles []LogFile
    Serverurl string //Client mode - URL to ship log to or to send command to be executed
    Logdbpath string
    Dbtimeout string
    Sslcert string
    Sslkey string
    PasswordFilterPatterns []string `yaml:"passwordfilterpatterns"`
    AppGoogleClientID string
    AppGoogleClientSecret string
    Sessionkey string
    AuthorizedDomain map[string]bool
    AWSLogs []AWSLogConfig
    IgnoreCertificateCheck bool
    LogRetention string
}

//AWSLogConfig -
type AWSLogConfig struct {
    LoggroupName string
    StreamPrefix []string
    FilterPtn string
    Profile string
    Region string
    Period string
}

//Config - Global
var Config AppConfig

//TimeISO8601LayOut
const (
    TimeISO8601LayOut = "2006-01-02T15:04:05-0700"
    AUTimeLayout = "02/01/2006 15:04:05 MST"
)

//GetDBConn -
func GetDBConn() (*sqlite3.Conn) {
    conn, err := sqlite3.Open(Config.Logdbpath)
	if err != nil {
		log.Fatalf("ERROR - can not open log database file - %v\n", err)
    }
	_dur, err := time.ParseDuration(Config.Dbtimeout)
	if err != nil {
        log.Printf("WARN - can not parse Dbtimeout string. Set default to 15 sec - %v\n", err)
        conn.BusyTimeout(15 * time.Second)
	} else{
		conn.BusyTimeout(_dur)
    }
    return conn
}

//GenerateDefaultConfig -
func GenerateDefaultConfig(opt ...interface{}) (e error) {
    defaultConfig := AppConfig {
        Port: 8000,
        Serverdomain: "localhost",
        Serverurl: "http://localhost:8000",
        IgnoreCertificateCheck: false,
        JwtKey: "ChangeThisKeyInYourSystem",
        Logdbpath: "logs.db",
        LogRetention: "90d",
        Dbtimeout: "45s",
        Commands: []Command {
            {
                Name: "example_ls",
                Path: "/bin/ls",
            },
            {
                Name: "ping",
                Path: "/bin/echo pong",
            },
        },
        Logfiles: []LogFile{
            {
                Name: "syslog",
                Paths: []string {
                    "/var/log/syslog", "/var/log/authlog", "/var/log/kern.log",
                },
                Timelayout: "Jan 02 15:04:05 2006 MST",
                Timepattern: `^([a-zA-Z]{3,3}[\s]+[\d]{0,2}[\s]+[\d]{2,2}\:[\d]{2,2}\:[\d]{2,2}) `,
                Timeadjust: "syslog",
                Timestrreplace: []string{",", "."},
                Pattern: `([^\s]+) ([^\s]+) (.*)$`,
                Multilineptn: `([^\s]+.*)$`,
                Excludepatterns: []string{},
                Includepatterns: []string{},
                Appname: "",
            },
        },
        Sslcert: "",
        Sslkey: "",
        PasswordFilterPatterns: []string {`([Pp]assword|[Pp]assphrase)['"]*[\:\=]*[\s\n]*[^\s]+[\s;]`},
    }

    var fPath, serverurl, jwtkey, logfile, appname, sslcert, sslkey string

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
            case "sslcert":
                sslcert = opt[i+1].(string)
            case "sslkey":
                sslkey = opt[i+1].(string)
            }
        }
    }

    var data []byte
    if logfile != "" {
        var logfiles []string
        _logfiles := strings.Split(logfile, ",")
        for _, _f := range(_logfiles) {
            if _, err := os.Stat(_f); os.IsNotExist(err) {
                log.Printf("INFO - File %s does not exist. In SimpleTail mode we dont wait, Skipping\n", _f)
            } else { logfiles = append(logfiles, _f) }
        }
        _Logfiles := []LogFile{
            {
                Name: "SimpleTailLog",
                Paths: logfiles,
                Timelayout: "Jan 02 15:04:05 2006 MST",
                Timepattern: "",
                Timeadjust: "",
                Pattern: `([^\s]+.*)$`,
                Multilineptn: `^[\s]+([^\s]+.*)$`,
                Excludepatterns: []string{},
                Includepatterns: []string{},
                Appname: appname,
            },
        }

        defaultConfig.Logfiles = _Logfiles
        defaultConfig.Serverurl = serverurl
        defaultConfig.JwtKey = jwtkey
        defaultConfig.IgnoreCertificateCheck = true

        data, e = yaml.Marshal(defaultConfig)
        if e != nil { log.Fatalf("ERROR can not dump default config yaml")}

    } else {
        defaultConfig.Sslcert, defaultConfig.Sslkey = sslcert, sslkey
        data, e = yaml.Marshal(defaultConfig)
        if e != nil { log.Fatalf("ERROR can not dump default config yaml")}
    }
    err := ioutil.WriteFile(fPath, data, 0600)
    if err != nil {return err}
    return LoadConfig(fPath)
}

// Scopes: OAuth 2.0 scopes provide a way to limit the amount of access that is granted to an access token.
var GoogleOauthConfig oauth2.Config

const OauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

//SessionStore -
var SessionStore *sessions.CookieStore

//ServerProtocol -
var ServerProtocol string

//CurrentYear - CurrentZone - Used for timeadjust
var CurrentYear, CurrentZone string

var PasswordFilterPtns []*regexp.Regexp

//LoadConfig -
func LoadConfig(fPath string) (e error) {
    yamlStr, e := ioutil.ReadFile(fPath)
    if e != nil {
        return e
    }
    e = yaml.Unmarshal(yamlStr, &Config)

    _t := time.Now()
    CurrentYear = strconv.FormatInt(int64(_t.Year()), 10)
    CurrentZone, _ = _t.Zone()

    for _, ptn := range(Config.PasswordFilterPatterns){
        PasswordFilterPtns = append(PasswordFilterPtns, regexp.MustCompile(ptn))
    }

    if Config.Sslkey == "" {
        ServerProtocol = "http"
    } else {
        ServerProtocol = "https"
    }

    GoogleOauthConfig = oauth2.Config {
        RedirectURL:  fmt.Sprintf("%s://%s:%d/auth/google/callback", ServerProtocol, Config.Serverdomain, Config.Port),
        // ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
        ClientID:     Config.AppGoogleClientID,
        ClientSecret: Config.AppGoogleClientSecret,
        Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
        Endpoint:     google.Endpoint,
    }
    SessionStore = sessions.NewCookieStore([]byte(Config.Sessionkey))
    SessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 4,
		HttpOnly: true,
    }

    return e
}