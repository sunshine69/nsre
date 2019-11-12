package cmd

import(
	"testing"
)

func TestTail(t *testing.T) {
	TailLog( &TailLogConfig{
			Path: "/var/log/syslog",
			Timelayout: "Jan 02 15:04:05 2006 MST",
			Timepattern: `([a-zA-Z]{3,3} [\d]{0,2} [\d]{2,2}\:[\d]{2,2}\:[\d]{2,2}) `,
			Timeadjust: "2019 AEST",
			Pattern: `([^\s]+) ([^\s]+) (.*)$`,
			SeekOffset: 0,
			Ekanitehost: "localhost:5514",
		},
	)
}