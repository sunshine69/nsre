package cmd

import (
	"net/http"
	"io/ioutil"
	"bytes"
	"github.com/bvinc/go-sqlite-lite/sqlite3"
	"os"
	"regexp"
	"log"
	"compress/zlib"
	"encoding/base64"
	"fmt"
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
	if tz == "" {
		tz, _ = time.Now().Zone()
	}
	timerangePtn := regexp.MustCompile(`([\d]{2,2}/[\d]{2,2}/[\d]{4,4} [\d]{2,2}:[\d]{2,2}:[\d]{2,2}) - ([\d]{2,2}/[\d]{2,2}/[\d]{4,4} [\d]{2,2}:[\d]{2,2}:[\d]{2,2})`)
	dur, e := time.ParseDuration(durationStr)
	if e != nil {
		log.Printf("ERROR can not parse duration string using time.ParseDuration for %s - %v. Will try next\n", durationStr, e)
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
//DecodeJenkinsConsoleNote - See https://github.com/LarrysGIT/Extract-Jenkins-Raw-Log/blob/master/README.md. Testing
func DecodeJenkinsConsoleNote(msg string) (string) {
	res := JenkinsLogDataPattern.FindStringSubmatch(msg)
	if res == nil { return msg }
	//See https://stackoverflow.com/questions/26635416/how-to-decrypt-jenkins-8mha-values the second answer for more info
	base64Data := res[1]
	decoded, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		fmt.Println("ERROR DecodeJenkinsConsoleNote base64decode error:", err)
		return msg
	}
	//trim first 40 bytes
	decoded = decoded[40:]
	// fmt.Printf("%s\n", decoded)
	var result = ""
	if r, e := zlib.NewReader(bytes.NewReader(decoded)); e == nil {
		if result1, e := ioutil.ReadAll(r); e == nil{
			result = string(result1)
		}
	}
	msg = JenkinsLogDataPattern.ReplaceAllString(msg, result)
	return msg
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
			message := FilterPassword(*msgStr, PasswordFilterPtns)
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

//ReadUserIP - parse the userIP:port from the request
func ReadUserIP(r *http.Request) string {
    IPAddress := r.Header.Get("X-Real-Ip")
    if IPAddress == "" {
        IPAddress = r.Header.Get("X-Forwarded-For")
    }
    if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
    return IPAddress
}

func Ternary(cond bool, first, second interface{}) interface{} {
	if cond {
		return first
	} else {
		return second
	}
}