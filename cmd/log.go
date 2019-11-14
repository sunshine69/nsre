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

	// seek := &tail.SeekInfo{Offset: 0, Whence: 0}
	// cfg.TailConfig.Location = seek

	for _, logFile := range(cfg.Paths) {
		log.Printf("Start tailing %s\n", logFile)
		t, e := tail.TailFile(logFile, cfg.TailConfig)
		if e != nil {
			log.Fatalf("Can not tail file - %v\n", e)
		}

		previousPos := LoadTailPosition(t, cfg)
		seek := &tail.SeekInfo{Offset: previousPos, Whence: 0}
		cfg.TailConfig.Location = seek

		if cfg.TailConfig.Follow {
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
		} else {
			ProcessTailLines(cfg, t)
			wg.Done()
		}
	}
}

//SaveTailPosition -
func SaveTailPosition(t *tail.Tail, cfg *TailLogConfig) {
	pos, e := t.Tell()
	if e != nil {
		log.Printf("Can not tell from tail where are we - %v\n", e)
	} else {
		filename := filepath.Join(os.Getenv("HOME"), "taillog-" + cfg.Name + "-" + filepath.Base(t.Filename))
		_pos := strconv.FormatInt(pos, 10)
		if e = ioutil.WriteFile(filename, []byte(_pos), 0750); e != nil {
			log.Printf("ERROR Can not save pos to %s - %v\n",filename ,e)
		}
	}
}

//LoadTailPosition -
func LoadTailPosition(t *tail.Tail, cfg *TailLogConfig) (int64) {
	filename := filepath.Join(os.Getenv("HOME"), "taillog-" + cfg.Name + "-" + filepath.Base(t.Filename))
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
		// fmt.Printf("EOF reached\n")
		return true
	}
	return false
}

func filterPassword(text string, passPtn *regexp.Regexp) (string) {
	return passPtn.ReplaceAllString(text, "$1 SENSITIVE_DATA_FILTERED ")
}

//SendLine -
func SendLine(timeParsed time.Time, hostStr, appNameStr, msgStr string, passPtn *regexp.Regexp) {
	logData := LogData{
		Timestamp: time.Now().UnixNano(),
		Datelog: timeParsed.UnixNano(),
		Host: hostStr,
		Application: appNameStr,
		Message: filterPassword(msgStr, passPtn),
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
	req, _ := http.NewRequest("POST", strings.Join([]string{Config.Serverurl, "log"}, "/"), bytes.NewBuffer(output))
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
	var timePtn, linePtn, multiLinePtn *regexp.Regexp
	linePtnStr := strings.Join([]string{cfg.Timepattern, cfg.Pattern}, "" )
	linePtn = regexp.MustCompile(linePtnStr)
	multiLinePtn = regexp.MustCompile(cfg.Multilineptn)
	passPtn := regexp.MustCompile(`([Pp]assword|[Pp]assphrase)['"]*[\:\=]*[\s\n]*[^\s]+[\s]`)

	if cfg.Timepattern != "" {
		timePtn = regexp.MustCompile(cfg.Timepattern)
		log.Printf("time ptn: '%s'\nline ptn: '%s'\n", cfg.Timepattern, linePtnStr)
	}

	timeLayout := cfg.Timelayout

	var timeParsed time.Time
	var e error
	var hostStr, appNameStr string

	if cfg.Appname != "" {
		appNameStr = cfg.Appname
	} else {
		appNameStr = "-"
	}

	lineStack := []string{}
	beginLineMatch := false

	for line := range tailLines {
		if line.Text == "" || line.Text == "\n" { continue }
		// fmt.Printf("Processing LineText: '%s'\n", line.Text)
		curSeek, _ := tail.Tell()
		if IsEOF(tail.Filename, curSeek) {
			msgStr := strings.Join(lineStack, "\n")
			lineStack = lineStack[:0]
			// log.Printf("EOF reached. Flush stack\n")
			SendLine(timeParsed, hostStr, appNameStr, msgStr, passPtn)
		}

		match := []string{"notimeptn"}
		if timePtn != nil {
			match = timePtn.FindStringSubmatch(line.Text)
		}

		if len(match) > 0 {
			beginLineMatch = true
			if len(lineStack) > 0 {//Flush the multiline stack
				msgStr := strings.Join(lineStack, "\n")
				lineStack = lineStack[:0]
				SendLine(timeParsed, hostStr, appNameStr, msgStr, passPtn)
			}
			if match[0] != "notimeptn" {
				timeStr := strings.Join([]string{match[1], cfg.Timeadjust}, " ")
				timeStr = strings.Replace(timeStr, "  ", " 0", -1)
				timeParsed, e = time.Parse(timeLayout, timeStr)
				if e != nil {
					log.Fatalf("ERROR Fail to parse time %v\n", e)
				}
			} else {
				timeParsed = time.Now()
			}

			match1 := linePtn.FindStringSubmatch(line.Text)
			matchCount := len(match1)
			if matchCount > 0 {
				var msgStr string
				switch matchCount {
				case 2: //no timePtn
					curHostname, _ := os.Hostname()
					hostStr, msgStr = curHostname, match1[1]
				case 3://For simple type we only support up to two matches, to parse the
					if match[0] == "notimeptn" {
						hostStr, msgStr = match1[1], match1[2]
					} else {
						curHostname, _ := os.Hostname()
						hostStr, msgStr = curHostname, match1[2]
					}
				case 4:
					hostStr, msgStr = match1[2], match1[3]
				case 5:
					hostStr, appNameStr, msgStr = match1[2], match1[3], match1[4]
				}
				if len(lineStack) == 0 {
					lineStack = append(lineStack, msgStr)
				}
			} else {
				// log.Printf("The pattern does not parse correct components. You need to have capture groups - TIMESTAMP HOSTNAME APP-NAME MSG\nLinePtn: '%s'.\n", linePtnStr)
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
				fmt.Printf("Line Text: '%s'\nPattern: %s\n", line.Text, cfg.Timepattern)
			}
		}
	}
}