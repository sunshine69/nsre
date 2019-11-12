package cmd

import (
	"sync"
	"strings"
	"bytes"
	"io/ioutil"
	"net/http"
	"time"
	"regexp"
	// "syscall"
	// "os/signal"
	"os"
	"log"
	"fmt"
	"github.com/hpcloud/tail"
	"github.com/json-iterator/go"
)

// type TailConfig tail.Config

//TailLogConfig -
type TailLogConfig struct {
    LogFile
	SeekOffset int64
	TailConfig tail.Config
}

//TailLog -
func TailLog(cfg *TailLogConfig, wg *sync.WaitGroup){
	// offset − This is the position of the read/write pointer within the file.
	// whence − This is optional and defaults to 0 which means absolute file positioning, other values are 1 which means seek relative to the current position and 2 means seek relative to the file's end.

	seek := &tail.SeekInfo{Offset: cfg.SeekOffset, Whence: 0}
	cfg.TailConfig.Location = seek

	// log.Printf("Start tailling config  %v\n", cfg)
	t, e := tail.TailFile(cfg.Path, cfg.TailConfig)
	if e != nil {
		log.Fatalf("Can not tail file - %v\n", e)
	}
	// c := make(chan os.Signal)
	// signal.Notify(c,
	// 	syscall.SIGHUP,
	// 	syscall.SIGINT,
	// 	syscall.SIGTERM,
	// 	syscall.SIGQUIT)

	ProcessLines(cfg, t.Lines)
	// s := <-c
	// log.Print(s.String())
	// t.Stop()
	wg.Done()
}
//ProcessLines -
func ProcessLines(cfg *TailLogConfig, tailLines chan *tail.Line) {

	timePtn := regexp.MustCompile(cfg.Timepattern)
	linePtnStr := fmt.Sprintf("%s%s", cfg.Timepattern, cfg.Pattern)
	linePtn := regexp.MustCompile(linePtnStr)
	log.Printf("time ptn: '%s'\nline ptn: '%s'\n", cfg.Timepattern, linePtnStr)

	timeLayout := cfg.Timelayout

	// lineStack := []string{}
	var timeParsed time.Time
	var e error

	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	client := &http.Client{}
    validToken, err := GenerateJWT()
    if err != nil {
        fmt.Println("Failed to generate token")
	}

	for line := range tailLines {

		match := timePtn.FindStringSubmatch(line.Text)
		if len(match) > 0 {
			timeStr := fmt.Sprintf("%s %s", match[1], cfg.Timeadjust)
			timeStr = strings.Replace(timeStr, "  ", " ", -1)
			timeParsed, e = time.Parse(timeLayout, timeStr)
			if e != nil {
				log.Fatalf("ERROR Fail to parse time %v\n", e)
			}

			match1 := linePtn.FindStringSubmatch(line.Text)
			matchCount := len(match1)
			if matchCount > 0 {
				var hostStr, appNameStr, msgStr string
				switch matchCount {
				case 3:
					curHostname, _ := os.Hostname()
					hostStr, appNameStr, msgStr = curHostname, "-", match1[2]
				case 4:
					hostStr, appNameStr, msgStr = match1[2], "-", match1[3]
				case 5:
					hostStr, appNameStr, msgStr = match1[2], match1[3], match1[4]
				}

				logData := LogData{
					Timestamp: time.Now().UnixNano(),
					Datelog: timeParsed.UnixNano(),
					Host: hostStr,
					Application: appNameStr,
					Message: msgStr,
				}
				output, e := json.Marshal(&logData)
				if e != nil {
					log.Fatalf("ERROR - can not marshal json to output - %v\n", e)
				}

				req, _ := http.NewRequest("POST", fmt.Sprintf("%s/log", Config.Serverurl), bytes.NewBuffer(output))
				req.Header.Set("Token", validToken)
				req.Header.Set("Content-Type", "application/json")

				res, err := client.Do(req)
				if err != nil {
					fmt.Printf("Error: %v", err)
				}
				_, err = ioutil.ReadAll(res.Body)
				if err != nil {
					fmt.Println(err)
				}
			} else {
				log.Fatalf("The pattern does not parse correct components. You need to have capture groups - TIMESTAMP HOSTNAME APP-NAME MSG\n")
			}
		} else {
			fmt.Printf("Line Text: '%s'\n", line.Text)
			log.Printf("Can not parse the time pattern\n")
		}
	}
}