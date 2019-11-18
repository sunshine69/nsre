package main

import (
	"time"
	"path/filepath"
	"github.com/hpcloud/tail"
	"sync"
	"os"
	"log"
	"flag"
	"github.com/sunshine69/nsre/cmd"
)

func startTailServer(tailCfg tail.Config) {
	if len(cmd.Config.Logfiles) == 0 { return }
	var wg sync.WaitGroup
	for _, _logFile := range(cmd.Config.Logfiles) {
		if len(_logFile.Paths) == 0 { continue }
		_tailLogConfig := cmd.TailLogConfig{
			LogFile: _logFile,
			TailConfig: tailCfg,
		}
		log.Printf("Spawn tailling process ...\n")
		wg.Add(1)
		go cmd.TailLog(&_tailLogConfig, &wg)
	}
	wg.Wait()
}

func main() {

	defaultConfig :=  filepath.Join(os.Getenv("HOME"), ".nsre.yaml")
	configFile := flag.String("c", defaultConfig, "Config file, default %s"+ defaultConfig)
	mode := flag.String("m", "client", "run mode. Can be server|client|tail|tailserver|reset.\nserver - start nsca server and wait for command.\nclient - take another option -cmd which is the command to send to the server.\ntail - tail the log and send to the log server.\nreset - reset the config using default")
	cmdName := flag.String("cmd", "", "Command name")
	tailFollow := flag.Bool("tailf", false, "Tail mode follow")

	tailFile := flag.String("f", "", "Files (coma sep list if more than 1) to parse in tailSimple mode.\nIt will take a file and parse by lines. There is no time parser. Need another option -appname to insert the application name, and -f <file to parse>; -url <log store url>.\nThis will ignore all config together.")
	serverURL := flag.String("url", "", "Server uri to post log to in tailSimple mode")
	appName := flag.String("appname", "", "Application name in tailSimple mode")
	jwtkey := flag.String("jwtkey", "", "JWT API Key to talk to server")
	sslcert := flag.String("sslcert", "", "SSL certificate path")
	sslkey := flag.String("sslkey", "", "SSL key path")
	poll := flag.Bool("poll", false, "Use polling file for tail. Usefull for windows.")

	flag.Parse()

	e := cmd.LoadConfig(*configFile)

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
		log.Printf("INFO Can not read config file. %v\nGenerating new one\n", e)
		if generateDefaultConfig() != nil {
			log.Fatalf("ERROR can not generate config file %v\n", e)
		}
	}

	tailCfg := tail.Config{
		// Location:    seek,
		ReOpen:      *tailFollow,
		MustExist:   false,
		Poll:        *poll,
		Pipe:        false,
		Follow:      *tailFollow,
		MaxLineSize: 0,
	}

	switch *mode {
	case "server":
		cmd.StartServer()
	case "client":
		cmd.RunCommand(*cmdName)
	case "nagios":
		cmd.RunNagiosCheckCommand(*cmdName)
	case "tail":
		startTailServer(tailCfg)
	case "tailserver":
		go cmd.StartServer()
		time.Sleep(2 * time.Second)
		startTailServer(tailCfg)
	case "reset", "setup":
		files, _ := filepath.Glob(filepath.Join(os.Getenv("HOME"), "taillog*"))
		for _, f := range(files) {
			os.Remove(f)
		}
		log.Printf("Going to generate config...")
		if generateDefaultConfig() != nil {
			log.Fatalf("ERROR can not generate config file %v\n", e)
		}
	}
}
