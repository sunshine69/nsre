package cmd

import (
	"time"
)

//Time handling
const (
	millisPerSecond     = int64(time.Second / time.Millisecond)
	nanosPerMillisecond = int64(time.Millisecond / time.Nanosecond)
	nanosPerSecond      = int64(time.Second / time.Nanosecond)
)

//NsToTime - 
func NsToTime(ns int64) time.Time  {
	secs := ns/nanosPerSecond
	nanos := ns - secs * nanosPerSecond
	return time.Unix(secs, nanos)
}