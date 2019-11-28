package main

import (
	"io/ioutil"
	"syscall"
	"os/signal"
	"fmt"
	"time"
	"path/filepath"
	"github.com/nxadm/tail"
	"sync"
	"os"
	"log"
	"flag"
	"github.com/sunshine69/nsre/cmd"
)

func startTailServer(tailCfg tail.Config, wg *sync.WaitGroup, c chan struct{}) {
	if len(cmd.Config.Logfiles) == 0 { return }
	for _, _logFile := range(cmd.Config.Logfiles) {
		if len(_logFile.Paths) == 0 { continue }
		_tailLogConfig := cmd.TailLogConfig{
			LogFile: _logFile,
			TailConfig: tailCfg,
		}
		log.Printf("Spawn tailling process ...\n")
		wg.Add(1)
		go cmd.TailLog(&_tailLogConfig, wg, c)
	}
	wg.Done()
}

func main() {

	defaultConfig :=  filepath.Join(os.Getenv("HOME"), ".nsre.yaml")
	configFile := flag.String("c", defaultConfig, "Config file, default %s"+ defaultConfig)
	mode := flag.String("m", "client", "run mode. Can be server|client|tail|tailserver|tailtest|cloudwatchlog|reset.\nserver - start nsca server and wait for command.\nclient - take another option -cmd which is the command to send to the server.\ntail - tail the log and send to the log server.\nreset - reset the config using default")
	cmdName := flag.String("cmd", "", "Command name")
	tailFollow := flag.Bool("tailf", false, "Tail mode follow")

	tailFile := flag.String("f", "", "Files (coma sep list if more than 1) to parse in tailSimple mode.\nIt will take a file and parse by lines. There is no time parser. Need another option -appname to insert the application name, and -f <file to parse>; -url <log store url>.\nThis will ignore all config together.")
	serverURL := flag.String("url", "", "Server uri to post log to in tailSimple mode")
	appName := flag.String("appname", "", "Application name in tailSimple mode")
	jwtkey := flag.String("jwtkey", "", "JWT API Key to talk to server")
	sslcert := flag.String("sslcert", "", "SSL certificate path")
	sslkey := flag.String("sslkey", "", "SSL key path")
	poll := flag.Bool("poll", false, "Use polling file for tail. Usefull for windows.")
	version := flag.Bool("version", false, "Get build version")

	flag.Parse()

	e := cmd.LoadConfig(*configFile)
	if *version {
		fmt.Println(cmd.Version)
		os.Exit(0)
	}

	var generateDefaultConfig = func() (error) {
		return cmd.GenerateDefaultConfig(
			"file", *configFile,
			"serverurl", *serverURL,
			"jwtkey", *jwtkey,
			"logfile", *tailFile,
			"appname", *appName,
			"sslcert", *sslcert,
			"sslkey", *sslkey,
		)
	}

	if e != nil {
		log.Printf("INFO Can not read config file. %v\nBack up and Generating new one\n", e)
		content, e := ioutil.ReadFile(*configFile)
		if e != nil {
			log.Fatalf("ERROR can not read config for backup - %v\n", e)
		} else {
			if e := ioutil.WriteFile(*configFile + ".bak",[]byte(content) ,0600); e != nil {
				log.Fatalf("ERROR writing backup config file %v\n", e)
			}
		}
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

	var wg sync.WaitGroup
	c := make(chan os.Signal, 4)
	signal.Notify(c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	switch *mode {
	case "server":
		go cmd.StartServer()
		s := <-c
		log.Printf("%s captured. Do cleaning up\n", s.String())
	case "client":
		cmd.RunCommand(*cmdName)
	case "nagios":
		cmd.RunNagiosCheckCommand(*cmdName)
	case "tail":
		wg.Add(1)
		exitCh := make(chan struct{})
		go startTailServer(tailCfg, &wg, exitCh)
		s := <-c
		log.Printf("%s captured. Do cleaning up\n", s.String())
		exitCh <- struct{}{}
		wg.Wait()
	case "awslog":
		exitCh1 := make(chan struct{})
		wg.Add(1)
		go cmd.StartAllAWSCloudwatchLogPolling(&wg, exitCh1)
		s := <-c
		log.Printf("%s captured. Do cleaning up\n", s.String())
		exitCh1 <- struct{}{}
	case "tailserver":
		go cmd.StartServer()
		time.Sleep(2 * time.Second)
		exitCh1 := make(chan struct{})
		exitCh2 := make(chan struct{})
		wg.Add(1)
		go cmd.StartAllAWSCloudwatchLogPolling(&wg, exitCh1)
		wg.Add(1)
		startTailServer(tailCfg, &wg, exitCh2)
		s := <-c
		log.Printf("%s captured. Do cleaning up\n", s.String())
		exitCh1 <- struct{}{}
		exitCh2 <- struct{}{}
	case "reset", "setup":
		files, _ := filepath.Glob(filepath.Join(os.Getenv("HOME"), "taillog*"))
		for _, f := range(files) {
			os.Remove(f)
		}
		log.Printf("Going to generate config...")
		if generateDefaultConfig() != nil {
			log.Fatalf("ERROR can not generate config file %v\n", e)
		}
	case "tailtest":
		cmd.TestTailLog(tailCfg, *tailFile)
	case "cloudwatchlog":
		cmd.ParseAWSCloudWatchLogEvent(*appName)
	}
}
