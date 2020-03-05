package main

import (
	"io/ioutil"
	"syscall"
	"os/signal"
	"fmt"
	"time"
	"path/filepath"
	"github.com/nxadm/tail"
	"os"
	"log"
	"flag"
	"github.com/mileusna/crontab"
	"github.com/sunshine69/nsre/cmd"
)

func startTailServer(tailCfg tail.Config, c chan os.Signal) {
	if len(cmd.Config.Logfiles) == 0 { c<- syscall.SIGQUIT; return }
	count := 0
	for _, _logFile := range(cmd.Config.Logfiles) {
		if len(_logFile.Paths) == 0 { continue }
		_tailLogConfig := cmd.TailLogConfig{
			LogFile: _logFile,
			TailConfig: tailCfg,
		}
		count = count + 1
		log.Printf("Spawn tailling process number %d ...\n", count)
		go cmd.TailLog(&_tailLogConfig, c)
	}
	if count == 0 { c<- syscall.SIGQUIT; return }
}

func BackupConfig(configFile *string) {
	if _, e := os.Stat(*configFile); os.IsNotExist(e) {
		log.Printf("INFO Config does not exist. Not backing up\n")
	} else {
		content, e := ioutil.ReadFile(*configFile)
		if e != nil {
			log.Fatalf("ERROR can not read config for backup - %v\n", e)
		} else {
			if e := ioutil.WriteFile(*configFile + ".bak",[]byte(content) ,0600); e != nil {
				log.Fatalf("ERROR writing backup config file %v\n", e)
			}
		}
	}
}

func RunScheduleTasks() {
	ctab := crontab.New() // create cron table
    // AddJob and test the errors
    if err := ctab.AddJob("1 0 1 * *", cmd.DatabaseMaintenance); err != nil {
        log.Printf("WARN - Can not add maintanance job - %v\n", err)
	}

}

func main() {
	defaultConfig :=  filepath.Join(os.Getenv("HOME"), ".nsre.yaml")
	configFile := flag.String("c", defaultConfig, "Config file, default %s"+ defaultConfig)
	mode := flag.String("m", "client", "run mode. Can be server|client|tail|tailserver|tailtest|cloudwatchlog|reset.\nserver - start nsca server and wait for command.\nclient - take another option -cmd which is the command to send to the server.\ntail - tail the log and send to the log server.\nreset - reset the config using default")
	cmdName := flag.String("cmd", "", "Command name")
	tailFollow := flag.Bool("tailf", false, "Tail mode follow")

	tailFile := flag.String("f", "", "Files (coma sep list if more than 1) to parse in tailSimple mode.\nIt will take a file and parse by lines. There is no time parser. Need another option -appname to insert the application name, and -f <file to parse>; -url <log store url>.\nThis will ignore all config together.")

	var serverURL string
	flag.StringVar(&serverURL, "serverurl", "", "Remote nsre url to send request to. Used by shipping log or nagios command, etc..")
	flag.StringVar(&serverURL, "url", "", "Remote nsre url to send request to. Used by shipping log or nagios command, etc..")

	logdbpath := flag.String("db", "", "Path to the application database file. Default: logs.db")
	port := flag.Int("port", 8000, "Server Port to listen on. Default is 8000")
	serverdomain := flag.String("domain", "", "Server domain. Leave it empty it will listen on default ip")
	appName := flag.String("appname", "", "Application name in tailSimple mode")
	jwtkey := flag.String("jwtkey", "", "JWT API Key to talk to server")
	sslcert := flag.String("sslcert", "", "SSL certificate path")
	sslkey := flag.String("sslkey", "", "SSL key path")
	poll := flag.Bool("poll", false, "Use polling file for tail. Usefull for windows.")
	version := flag.Bool("version", false, "Get build version")

	flag.Parse()

	if *version {
		fmt.Println(cmd.Version)
		os.Exit(0)
	}

	if (*mode == "client") && (*cmdName == "") {
		log.Fatalf("Mode client require option -cmd for command name\n. Run with option -h for help\n")
	}

	var generateDefaultConfig = func() (error) {
		return cmd.GenerateDefaultConfig(
			"file", *configFile,
			"serverurl", serverURL,
			"jwtkey", *jwtkey,
			"logfile", *tailFile,
			"appname", *appName,
			"sslcert", *sslcert,
			"sslkey", *sslkey,
			"port", *port,
			"serverdomain", *serverdomain,
			"logdbpath", *logdbpath,
		)
	}

	if e := cmd.LoadConfig(*configFile); e!= nil {
		log.Printf("INFO Can not read config file. %v\nBack up and Generating new one\n", e)
		BackupConfig(configFile)
		if generateDefaultConfig() != nil {
			log.Fatalf("ERROR can not generate config file %v\n", e)
		}
	}

	seek := tail.SeekInfo{Offset: 0, Whence: 0}
	tailCfg := tail.Config{
		Location:    &seek,
		ReOpen:      *tailFollow,
		MustExist:   false,
		Poll:        *poll,
		Pipe:        false,
		Follow:      *tailFollow,
		MaxLineSize: 0,
	}

	c := make(chan os.Signal, 4)
	signal.Notify(c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	switch *mode {
	case "server":
		RunScheduleTasks()
		go cmd.StartServer()
		s := <-c
		log.Printf("%s captured. Do cleaning up\n", s.String())
	case "client":
		cmd.RunCommand(*cmdName)
	case "nagios":
		cmd.RunNagiosCheckCommand(*cmdName)
	case "tail":
		exitCh := make(chan os.Signal)
		go startTailServer(tailCfg, exitCh)
		if *tailFollow{
			s := <-c
			log.Printf("%s captured. Do cleaning up\n", s.String())
			exitCh<- s
			s = <-exitCh
		} else { <-exitCh }
	case "awslog":
		exitCh1 := make(chan os.Signal)
		go cmd.StartAllAWSCloudwatchLogPolling(exitCh1)
		s := <-c
		log.Printf("%s captured. Do cleaning up\n", s.String())
		exitCh1 <- s
		s = <-exitCh1
	case "tailserver":
		RunScheduleTasks()
		go cmd.StartServer()
		time.Sleep(2 * time.Second)
		exitCh1 := make(chan os.Signal)
		exitCh2 := make(chan os.Signal)
		go cmd.StartAllAWSCloudwatchLogPolling(exitCh1)
		go startTailServer(tailCfg, exitCh2)
		s := <-c
		log.Printf("%s captured. Do cleaning up\n", s.String())
		exitCh1 <- s
		exitCh2 <- s
		s = <-exitCh1
		s = <-exitCh2
	case "reset", "setup":
		files, _ := filepath.Glob(filepath.Join(os.Getenv("HOME"), "taillog*"))
		for _, f := range(files) {
			os.Remove(f)
		}
		log.Printf("Going to generate config...")
		BackupConfig(configFile)
		if generateDefaultConfig() != nil {
			log.Fatalf("ERROR can not generate config file %s\n", *configFile)
		}
	case "tailtest":
		cmd.TestTailLog(tailCfg, *tailFile)
	case "cloudwatchlog":
		cmd.ParseAWSCloudWatchLogEvent(*appName)
	}
}
