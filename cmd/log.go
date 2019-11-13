package cmd

import (
	"strconv"
	"path/filepath"
	"io"
	"sync"
	"strings"
	"bytes"
	"io/ioutil"
	"net/http"
	"time"
	"regexp"
	"syscall"
	"os/signal"
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

	previousPos := LoadTailPosition(cfg)
	seek := &tail.SeekInfo{Offset: previousPos, Whence: 0}
	cfg.TailConfig.Location = seek

	for _, logFile := range(cfg.Paths) {
		t, e := tail.TailFile(logFile, cfg.TailConfig)
		if e != nil {
			log.Fatalf("Can not tail file - %v\n", e)
		}
		c := make(chan os.Signal, 4)
		signal.Notify(c,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		go ProcessTailLines(cfg, t)
		s := <-c
		log.Printf("%s captured. Do cleaning up\n", s.String())
		SaveTailPosition(t, cfg)
		t.Stop()
		wg.Done()
	}
}

//SaveTailPosition -
func SaveTailPosition(t *tail.Tail, cfg *TailLogConfig) {
	pos, e := t.Tell()
	if e != nil {
		log.Printf("Can not tell from tail where are we - %v\n", e)
	} else {
		filename := filepath.Join(os.Getenv("HOME"), "taillog-" + cfg.Name)
		_pos := strconv.FormatInt(pos, 10)
		if e = ioutil.WriteFile(filename, []byte(_pos), 0750); e != nil {
			log.Printf("ERROR Can not save pos to %s - %v\n",filename ,e)
		}
	}
}

//LoadTailPosition -
func LoadTailPosition(cfg *TailLogConfig) (int64) {
	filename := filepath.Join(os.Getenv("HOME"), "taillog-" + cfg.Name)
	data, e := ioutil.ReadFile(filename)
	if e != nil {
		log.Printf("ERROR Can not read previous pos. Will set seek to 0 - %s\n", e)
		// os.Remove(filename)
		return 0
	}
	out, e := strconv.ParseInt(string(data), 10, 64)
	if e != nil {
		log.Printf("ERROR Can not parse int previous pos. Will set seek to 0 - %s\n", e)
		return 0
	}
	log.Printf("Loaded previous file pos %d from %s. To set from beginnng remove the file\n", out, filename)
	return out
}


//IsEOF - NOt sure why tail does not provide this test.
func IsEOF(filename string, seek int64) (bool) {
	fh, e := os.Open(filename)
	defer fh.Close()
	if e != nil {
		fmt.Printf("ERROR can not open file - %v\n", e)
	}
	buff := make([]byte, 1)
	fh.Seek(seek, 0)
	_, e = fh.Read(buff)
	if e == io.EOF {
		fmt.Printf("ERROR\n")
		return true
	}
	return false
}

//SendLine -
func SendLine(timeParsed time.Time, hostStr, appNameStr, msgStr string) {
	logData := LogData{
		Timestamp: time.Now().UnixNano(),
		Datelog: timeParsed.UnixNano(),
		Host: hostStr,
		Application: appNameStr,
		Message: msgStr,
	}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	output, e := json.Marshal(&logData)
	if e != nil {
		log.Fatalf("ERROR - can not marshal json to output - %v\n", e)
	}
	client := &http.Client{}
    validToken, err := GenerateJWT()
    if err != nil {
        fmt.Println("Failed to generate token")
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
}

//ProcessTailLines -
func ProcessTailLines(cfg *TailLogConfig, tail *tail.Tail) {
	tailLines := tail.Lines
	timePtn := regexp.MustCompile(cfg.Timepattern)
	linePtnStr := fmt.Sprintf("%s%s", cfg.Timepattern, cfg.Pattern)
	linePtn := regexp.MustCompile(linePtnStr)
	multiLinePtn := regexp.MustCompile(cfg.Multilineptn)
	log.Printf("time ptn: '%s'\nline ptn: '%s'\n", cfg.Timepattern, linePtnStr)

	timeLayout := cfg.Timelayout

	var timeParsed time.Time
	var e error
	var hostStr, appNameStr string

	lineStack := []string{}
	beginLineMatch := false

	for line := range tailLines {
		curSeek, _ := tail.Tell()
		if IsEOF(tail.Filename, curSeek) {
			msgStr := strings.Join(lineStack, "\n")
			lineStack = lineStack[:0]
			// log.Printf("EOF reached. Flush stack\n")
			SendLine(timeParsed, hostStr, appNameStr, msgStr)
		}
		match := timePtn.FindStringSubmatch(line.Text)
		if len(match) > 0 {
			beginLineMatch = true
			if len(lineStack) > 0 {//Flush the multiline stack
				msgStr := strings.Join(lineStack, "\n")
				lineStack = lineStack[:0]
				SendLine(timeParsed, hostStr, appNameStr, msgStr)
			}
			timeStr := fmt.Sprintf("%s %s", match[1], cfg.Timeadjust)
			timeStr = strings.Replace(timeStr, "  ", " ", -1)
			timeParsed, e = time.Parse(timeLayout, timeStr)
			if e != nil {
				log.Fatalf("ERROR Fail to parse time %v\n", e)
			}
			match1 := linePtn.FindStringSubmatch(line.Text)
			matchCount := len(match1)
			if matchCount > 0 {
				var msgStr string
				switch matchCount {
				case 3:
					curHostname, _ := os.Hostname()
					hostStr, appNameStr, msgStr = curHostname, "-", match1[2]
				case 4:
					hostStr, appNameStr, msgStr = match1[2], "-", match1[3]
				case 5:
					hostStr, appNameStr, msgStr = match1[2], match1[3], match1[4]
				}
				if len(lineStack) == 0 {
					lineStack = append(lineStack, msgStr)
				}
			} else {
				log.Fatalf("The pattern does not parse correct components. You need to have capture groups - TIMESTAMP HOSTNAME APP-NAME MSG\n")
			}
		} else {
			if beginLineMatch {
				mMatch := multiLinePtn.FindStringSubmatch(line.Text)
				if len(mMatch) > 0 {
					if len(lineStack) > 0 {
						// fmt.Printf("Part of multiLine Text: '%s'\n", line.Text)
						lineStack = append(lineStack, mMatch[1])
					}
				} else {
					beginLineMatch = false
					log.Printf("Can not parse multiline pattern\n")
					fmt.Printf("Line Text: '%s'\n", line.Text)
				}
			} else {
				log.Printf("Can not parse time pattern\n")
				fmt.Printf("Line Text: '%s'\n", line.Text)
			}
		}
	}
}