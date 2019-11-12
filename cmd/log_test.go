package cmd

import(
	"testing"
	"fmt"
	"os"
)

func TestTail(t *testing.T) {
	defaultConfig :=  fmt.Sprintf("%s/.nsca-go.yaml", os.Getenv("HOME"))
	LoadConfig(defaultConfig)

	logFile := LogFile {
		Path: "/var/log/syslog",
		Timelayout: "Jan 02 15:04:05 2006 MST",
		Timepattern: `([a-zA-Z]{3,3} [\d]{0,2} [\d]{2,2}\:[\d]{2,2}\:[\d]{2,2}) `,
		Timeadjust: "2019 AEST",
		Pattern: `([^\s]+) ([^\s]+) (.*)$`,
	}
	TailLog( &TailLogConfig{
			LogFile: logFile,
			SeekOffset: 0,
		},
	)
}