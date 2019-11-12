package cmd

import (

	"net"
	"time"
	"regexp"
	// "syscall"
	// "os/signal"
	// "os"
	"log"
	"fmt"
	"github.com/hpcloud/tail"

)
//TailLogConfig -
type TailLogConfig struct {
    Path string
    Timelayout string //Parse the match below into go time object
    Timepattern string //extract the timestamp part into a timeStr which is fed into the Timelayout
	Timeadjust string //If the time extracted string miss some info (like year or zone etc) this string will be appended to the string
	Pattern string //will be matched to extract the HOSTNAME APP-NAME PROC-ID MSGID MSG part of the line.
	SeekOffset int64
	Ekanitehost string
}

//TailLog -
func TailLog(cfg *TailLogConfig){
	// offset − This is the position of the read/write pointer within the file.
	// whence − This is optional and defaults to 0 which means absolute file positioning, other values are 1 which means seek relative to the current position and 2 means seek relative to the file's end.

	seek := &tail.SeekInfo{Offset: cfg.SeekOffset, Whence: 0}

	tailConfig := tail.Config {
		Location:    seek,
		ReOpen:      true,
		MustExist:   false,
		Poll:        false,
		Pipe:        false,
    	Follow:      true,
    	MaxLineSize: 0,
	}

	t, e := tail.TailFile(cfg.Path, tailConfig)
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

	conn, e := net.Dial("tcp", cfg.Ekanitehost)
	if e != nil {
		log.Fatalf("ERROR can not dial to host %s\n", cfg.Ekanitehost)
	}


	for line := range tailLines {

		match := timePtn.FindStringSubmatch(line.Text)
		if len(match) > 0 {
			timeStr := fmt.Sprintf("%s %s", match[1], cfg.Timeadjust)
			timeParsed, e = time.Parse(timeLayout, timeStr)
			if e != nil {
				log.Printf("ERROR Fail to parse time\n")
			}
			timeStr = timeParsed.Format(TimeISO8601LayOut)
			match1 := linePtn.FindStringSubmatch(line.Text)
			log.Printf("Matched groups: %d\n", len(match1))
			if len(match1) > 0 {
				hostStr, appNameStr, procIDStr, msgIDStr, msgStr := match1[2], match1[3], "-", "-", match1[4]
				output := fmt.Sprintf("<134>0 %s %s %s %s %s %s", timeStr, hostStr, appNameStr, procIDStr, msgIDStr, msgStr)
				fmt.Printf("Going to Send Line Text: '%s'\n", output)
				fmt.Fprintf(conn, output + "\n")
				// message, _ := bufio.NewReader(conn).ReadString('\n')
				// log.Print("Message from server: "+message)
			} else {
				log.Fatalf("The pattern does not parse into 6 component. You need to have 6 capture groups - TIMESTAMP HOSTNAME APP-NAME PROC-ID MSGID MSG\n")
			}
		} else {
			fmt.Printf("Line Text: '%s'\n", line.Text)
			log.Fatalf("Can not parse the time pattern\n")
		}
	}
}