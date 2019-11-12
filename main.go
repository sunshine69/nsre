package main

import (
	"os"
	"fmt"
	"log"
	"flag"
	"./cmd"
)

func main() {
	defaultConfig :=  fmt.Sprintf("%s/.nsca-go.yaml", os.Getenv("HOME"))
	configFile := flag.String("c", defaultConfig, fmt.Sprintf("Config file, default %s", defaultConfig))
	mode := flag.String("m", "client", "run mode. Can be server|client|tail|reset.\nserver - start nsca server and wait for command.\nclient - take another option -cmd which is the command to send to the server.\ntail - tail the log and send to the log server.\nreset - reset the config using default")
	cmdName := flag.String("cmd", "", "Command name")
	flag.Parse()

	e := cmd.LoadConfig(*configFile)
	if e != nil {
		log.Printf("Error reading config file. %v\nGenerating new one\n", e)
		if e = cmd.GenerateDefaultConfig(*configFile); e != nil {
			log.Fatalf("ERROR can not geenrate config file %v\n", e)
		}
	}
	switch *mode {
	case "server":
		cmd.StartServer()
	case "client":
		cmd.SendRequest(*cmdName)
	case "tail":
		log.Printf("mode tail\n")
	case "reset":
		cmd.GenerateDefaultConfig(*configFile)
	}
}
