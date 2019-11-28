package cmd

import (
	"github.com/bvinc/go-sqlite-lite/sqlite3"
	"os"
	"regexp"
	"log"
	// "fmt"
	// "io"
	// "compress/gzip"
	// "bytes"
	// "encoding/base64"
	"strings"
	"encoding/hex"
	"crypto/md5"
	"time"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

//Time handling
const (
	MillisPerSecond     = int64(time.Second / time.Millisecond)
	NanosPerMillisecond = int64(time.Millisecond / time.Nanosecond)
	NanosPerSecond      = int64(time.Second / time.Nanosecond)
)

//NsToTime -
func NsToTime(ns int64) time.Time {
	secs := ns/NanosPerSecond
	nanos := ns - secs * NanosPerSecond
	return time.Unix(secs, nanos)
}

//MsToTime -
func MsToTime(ms int64) time.Time {
	secs := ms/MillisPerSecond
	nanos := (ms - secs * MillisPerSecond) * NanosPerMillisecond
	return time.Unix(secs, nanos)
}

//ParseTimeRange -
func ParseTimeRange(durationStr, tz string) (time.Time, time.Time) {
	var start, end time.Time
	timerangePtn := regexp.MustCompile(`([\d]{2,2}/[\d]{2,2}/[\d]{4,4} [\d]{2,2}:[\d]{2,2}:[\d]{2,2}) - ([\d]{2,2}/[\d]{2,2}/[\d]{4,4} [\d]{2,2}:[\d]{2,2}:[\d]{2,2})`)
	dur, e := time.ParseDuration(durationStr)
	if e != nil {
		log.Printf("ERROR can not parse durection string using time.ParseDuration for %s - %v. Will try next\n", durationStr, e)
		m := timerangePtn.FindStringSubmatch(durationStr)
		if len(m) != 3 {
			log.Printf("ERROR Can not parse duration. Set default to 15m ago - %v", e)
			dur, _ = time.ParseDuration("15m")
		} else {
			start, _ = time.Parse(AUTimeLayout, m[1] + " " + tz )
			end, _ = time.Parse(AUTimeLayout, m[2] + " " + tz)
		}
	} else {
		end = time.Now()
		start = end.Add(-1 * dur)
	}
	// log.Printf("Time range: %s - %s\n",start.Format(AUTimeLayout), end.Format(AUTimeLayout))
	return start, end
}

//CreateHash - md5
func CreateHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}
//DecodeJenkinsConsoleNote - See https://github.com/LarrysGIT/Extract-Jenkins-Raw-Log/blob/master/README.md. Not working yet TODO
func DecodeJenkinsConsoleNote(consoleNote string) (string) {
	return consoleNote
	// PREAMBLE_STR := "\u001B[8mha:"
	// POSTAMBLE_STR := "\u001B[0m"
	// PREAMBLE_STR := `[8mha:`
	// POSTAMBLE_STR := `[0m`
	// pos := strings.Index(PREAMBLE_STR, consoleNote)
	// fmt.Printf("Call DecodeJenkinsConsoleNote pos %d\n%s", pos, consoleNote)
	// if pos == -1 { return consoleNote }
	// pos1 := strings.Index(POSTAMBLE_STR, consoleNote)
	// if pos == -1 { return consoleNote }
	// posStartExtract := pos + len(PREAMBLE_STR)
	// posEndExtract := pos1 - len(POSTAMBLE_STR) - 1
	// data := consoleNote[posStartExtract:posEndExtract]
	// fmt.Print(data)
	// dataByte, err := base64.StdEncoding.DecodeString(data)
	// dataByte1 := dataByte[40:]
	// var buf bytes.Buffer
	// buf.Write(dataByte1)

	// zr, err := gzip.NewReader(&buf)
	// if err != nil {
	// 	fmt.Printf("ERROR Can not ungzip. Will return string as is - %v\n", err)
	// 	return consoleNote
	// }

	// var buf1 bytes.Buffer
	// if _, err := io.Copy(&buf1, zr); err != nil {
	// 	fmt.Printf("ERROR can not copy uncompressed data - %v\n", err)
	// }

	// if err := zr.Close(); err != nil {
	// 	fmt.Printf("ERROR can not close the gzip reader - %v\n", err)
	// }

	// return consoleNote[0:(pos-1)] + buf1.String() + consoleNote[(pos1 + 1):]
}

//CheckAuthorizedUser -
func CheckAuthorizedUser(email string) (bool) {
	conn := GetDBConn()
	defer conn.Close()
	q := "SELECT email FROM user where email = '" + email + "';"
	stmt, err := conn.Prepare(q)
	if err != nil {
		log.Printf("ERROR - %v\n", err)
	}
	defer stmt.Close()

	hasRow, err := stmt.Step()
	if !hasRow {
		return CheckAuthorizedDomain(email)
	} else {
		return true
	}
}

//CheckAuthorizedDomain -
func CheckAuthorizedDomain(email string) (bool) {
	_tmp := strings.Split(email, "@")
	ok, e := Config.AuthorizedDomain[_tmp[1]]
	return ok && e
}

//SendAWSLogEvents - Store the last End time in the event list
func SendAWSLogEvents(evts []*cloudwatchlogs.FilteredLogEvent, appNameStr string, timeMark int64, conn *sqlite3.Conn) (int64) {
	evtCount := len(evts)
	log.Printf("Events count: %d\n", evtCount)
	if evtCount == 0 { return 0 }

	var timeParsed time.Time

	hostStr, _ := os.Hostname()

	for idx, data := range(evts) {
		//Start include the previous record thus skip it except from beginning
		if (timeMark > 0) && (idx == 0) { continue }
		// log.Printf("Send ID: %s - time %d\n", *data.EventId, *data.Timestamp)
		timeHarvest := time.Now()
		timeParsed = MsToTime(*data.Timestamp)
		logFile, msgStr :=  data.LogStreamName, data.Message

		if conn != nil {
			message := FilterPassword(*msgStr, PasswordFilterPtn)
			err := conn.Exec(`INSERT INTO log(timestamp, datelog, host, application, logfile, message) VALUES (?, ?, ?, ?, ?, ?)`, timeHarvest.UnixNano(), timeParsed.UnixNano(), hostStr, appNameStr, *logFile, message)
			if err != nil {
				log.Printf("ERROR - can not insert data for logline - %v\n", err)
			}
		} else {
			SendLine(timeHarvest, timeParsed, hostStr, appNameStr, *logFile, *msgStr)
		}
	}
	return timeParsed.UnixNano() / NanosPerMillisecond
}