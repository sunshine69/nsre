package cmd

import (
	"time"
	"bytes"
	"fmt"
	"net/http/httputil"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

//HandleSNSEvent - From/To/action
func HandleSNSEvent(w http.ResponseWriter, r *http.Request) {
	msg, _ := ioutil.ReadAll(r.Body)
	SaveDumpData("HandleSNSEvent", "twilio_call", r.Method, string(msg))
	vars := mux.Vars(r)
	From, To, Action := vars["from"], vars["to"], vars["action"]
	Subj := json.Get(msg, "Subject").ToString()
	Subj = Ternary(Subj == "", "no subject", Subj).(string)
	Body := json.Get(msg, "Message").ToString()
	Body = Ternary(Body == "", string(msg), Body).(string)
	MakeTwilioCall("", Action, Body, From, To, "HandleSNSEvent", Subj, "")
}

// For debugging purposes only. The endpoint handler can temporary call this to examine the data structure
func SaveDumpData(host, application, logfile, message string) *LogData {
	host = Ternary(host == "", "DEBUG", host).(string)
	application = Ternary(application == "", "DUMPER", application).(string)

	logData := LogData{
		Timestamp: time.Now().UnixNano(),
		Datelog: time.Now().UnixNano(),
		Host: host,
		Application: application,
		Logfile: logfile,
		Message: message,
	}
	data, _ := json.Marshal(logData)
	InsertLog(data)
	return &logData
}

func DumpPost(w http.ResponseWriter, r *http.Request) {
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("DEBUG - DUMP\n\n%s\n",requestDump)
	msg, _ := ioutil.ReadAll(r.Body)
	fmt.Printf("DEBUG - Body\n\n%s\n",msg)

	SaveDumpData("", "", r.Method, string(msg))

	r.Body = ioutil.NopCloser(bytes.NewBuffer(msg))
	fmt.Fprintf(w, "OK")
	return
}