package cmd

import (
	"bytes"
	"strconv"
	"strings"
	"github.com/gorilla/mux"
	"time"
	"io/ioutil"
	"log"
	"net/url"
	"net/http"
	"fmt"
	"github.com/google/uuid"
	"github.com/json-iterator/go"
)

// Simple twilio for nagios call
// This to allow place call/sms with text to twillio. Twillio api call is a fire off thing, we need to query the state and handle it properly.

//This does not intend to be full featured. Instead it tries to keep simple and just to be used for nagios notification only

//This app will create a listener /twilio/status_callback to take the status call back from Twillio
// /twilio/call|sms - Make a call or sms
// It would use the existing LogData database to log the call state queue and re-try if failed state occured

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func ProcessTwilioCallEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	myCallId := vars["call_id"]

	currentItem := GetTwilioCall(myCallId)
	if currentItem == "" {//No previous call.
		return
	}
	rawMessage, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(rawMessage))//restore body for parseform
	r.ParseForm()

	//We only care about MessageStatus and CallStatus
	msg := fmt.Sprintf(`{
		"CallStatus": "%s",
		"MessageStatus": "%s",
		"ErrorCode": "%s",
		"RawMessage": "%s"
	}
	`, r.FormValue("CallStatus"), r.FormValue("MessageStatus"), r.FormValue("ErrorCode"), rawMessage)

	logData := LogData{
		Timestamp: time.Now().UnixNano(),
		Datelog: time.Now().UnixNano(),
		Host: "twilio_call",
		Application: myCallId,
		Logfile: "",
		Message: msg,
	}
	data, _ := json.Marshal(logData)
	InsertLog(data)
	fmt.Fprintf(w, "OK")
	return
}

func MakeTwilioCall(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reqAction := vars["action"]

	r.ParseForm()
	Body := r.FormValue("Body")
	From := r.FormValue("From")
	To := r.FormValue("To")
	fmt.Printf("DEBUG Body: %s - From: %s - To: %s\n", Body, From, To)
	twilioSid := GetConfig("twilio_sid")
	twilioSec := GetConfig("twilio_sec")
	//Twilio will post to this url + /<my_call_sid>
	twilioStatusCallBack := GetConfigSave("twilio_callback", "https://log.xvt.technology/twilio/events/")

	twilioCallUrl, Twiml := "", ""
	myCallId := uuid.New().String()
	formData := url.Values{}

	switch reqAction {
	case "call":
		twilioCallUrl = GetConfigSave("twilio_url", "https://api.twilio.com/2010-04-01/Accounts/" + twilioSid + "/Calls.json")
		Twiml = `<?xml version="1.0" encoding="UTF-8"?><Response><Say voice="alice">` + Body + `</Say></Response>`
		formData = url.Values{
			"Twiml": { Twiml },
			"From": { From },
			"To": { To },
			"StatusCallbackMethod": {"POST"},
			"StatusCallback": { twilioStatusCallBack + myCallId },
		}
	case "sms":
		twilioCallUrl = GetConfigSave("twilio_url", "https://api.twilio.com/2010-04-01/Accounts/" + twilioSid + "/Messages.json")
		formData = url.Values{
			"From": { From },
			"To": { To },
			"StatusCallbackMethod": {"POST"},
			"StatusCallback": { twilioStatusCallBack + myCallId },
		}
	}
	// twilioCallUrl = "https://note.xvt.technology:8000/dumppost"

	makeCall := func() {
		encodedData := formData.Encode()
		req, _ := http.NewRequest("POST", twilioCallUrl , strings.NewReader(encodedData))
		req.SetBasicAuth(twilioSid, twilioSec)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(encodedData)))

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			log.Printf("ERROR MakeTwilioCall Send Req %v\n", err)
			http.Error(w, fmt.Sprintf("ERROR %v", err), 500)
			return
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Printf("ERROR MakeTwilioCall Get Response %v\n", err)
			http.Error(w, fmt.Sprintf("ERROR %v", err), 500)
			return
		}

		callSid := json.Get(body, "sid").ToString()

		logData := LogData{
			Timestamp: time.Now().UnixNano(),
			Datelog: time.Now().UnixNano(),
			Host: "twilio_call",
			Application: myCallId,
			Logfile: callSid,
			Message: string(body),
		}

		data, _ := json.Marshal(logData)
		InsertLog(data)
	}
	AssertCall := func() {
		tryCount := 0
		action := ""
		for {//Call and re-call if call fail
			tryCount = tryCount + 1
			existingCall := GetTwilioCall(myCallId)
			fmt.Printf("DEBUG count: %d - existingCall '%s'\nAction: '%s'\n", tryCount, existingCall, action)
			if existingCall == "" || action == "make_call" { //New call
				makeCall()
			}
			CallStatus := json.Get([]byte(existingCall), "CallStatus").ToString()
			fmt.Printf("DEBUG CallStatus '%s'\n", CallStatus)
			switch CallStatus {
			case "completed":
				action = "exit"
				break
			case "ringing", "queued", "in-progress", "":
				action = "wait"
			case "busy", "failed", "no-answer":
				action = "make_call"
			}
			if action == "exit" { fmt.Printf("DEBUG call suceeded\n"); break }
			time.Sleep(15 * time.Second)
			if tryCount > 10 {
				log.Printf("INFO TryCount exeeded %d\n", 10)
				action = "fail"
				break
			}
		}
		return
	}
	go AssertCall()
	fmt.Fprintf(w, "OK scheduled")
	return
}

func GetTwilioCall(myCallId string) string {
	DB := GetDBConn(); defer DB.Close()
	start, end := ParseTimeRange("1h", "AEST")
	stmt, e := DB.Prepare(`SELECT message from log WHERE ((timestamp > ?) AND (timestamp < ?)) AND  application = ? ORDER BY timestamp DESC`, start.UnixNano(), end.UnixNano(), myCallId)
	if e != nil { fmt.Printf("ERROR %v\n", e) }
	defer stmt.Close()
	stmt.Step()
	var body string
	stmt.Scan(&body)
	return body
}