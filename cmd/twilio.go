package cmd

import (
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

	body, _ := ioutil.ReadAll(r.Body)

	logData := LogData{
		Timestamp: time.Now().UnixNano(),
		Datelog: time.Now().UnixNano(),
		Host: "twilio_call",
		Application: myCallId,
		Logfile: "",
		Message: string(body),
	}
	data, _ := json.Marshal(logData)
	InsertLog(data)
	fmt.Fprintf(w, "OK")
	return
}

func MakeTwilioCall(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	Body := r.FormValue("Body")
	From := r.FormValue("From")
	To := r.FormValue("To")
	fmt.Printf("DEBUG Body: %s - From: %s - To: %s\n", Body, From, To)
	twilioSid := GetConfig("twilio_sid")
	twilioSec := GetConfig("twilio_sec")
	//Twilio will post to this url + /<my_call_sid>
	twilioStatusCallBack := GetConfigSave("twilio_callback", "https://log.xvt.technology/twilio/events/")
	twilioCallUrl := GetConfigSave("twilio_url", "https://api.twilio.com/2010-04-01/Accounts/" + twilioSid + "/Calls.json")
	// twilioCallUrl = "https://note.xvt.technology:8000/dumppost"

	// twilioSmslUrl := GetConfigSave("twilio_url", "https://api.twilio.com/2010-04-01/Accounts/" + twilioSid + "/Messages.json")

	Twiml := `<?xml version="1.0" encoding="UTF-8"?><Response><Say voice="alice">` + Body + `</Say></Response>`

	myCallId := uuid.New().String()

	formData := url.Values{
		"Twiml": { Twiml },
		"From": { From },
		"To": { To },
		"StatusCallbackMethod": {"POST"},
		"StatusCallback": { twilioStatusCallBack + myCallId },
	}

	makeCall := func() {
		encodedData := formData.Encode()
		fmt.Printf("DEBUG %s\n", encodedData)
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
	tryCount := 0
	for {//Call and re-call if call fail
		tryCount = tryCount + 1
		existingCall := GetTwilioCall(myCallId)
		action := "make_call"
		if existingCall == "" && action == "make_call" { //New call
			makeCall()
		} else {
			status := json.Get([]byte(existingCall), "CallStatus").ToString()
			switch status {
			case "completed":
				action = "exit"
				break
			case "ringing", "queued", "in-progress":
				action = "wait"
			case "busy", "failed", "no-answer":
				action = "make_call"
			}
		}
		if action == "exit" { break }
		time.Sleep(15 * time.Second)
		if tryCount > 3 {
			log.Printf("INFO TryCount exeeded %d\n", tryCount)
			action = "fail"
			break
		}
	}
}

func GetTwilioCall(myCallId string) string {
	DB := GetDBConn(); defer DB.Close()
	stmt, _ := DB.Prepare(`SELECT message from log WHERE application = ? ORDER BY timestamp DESC`, myCallId); defer stmt.Close()
	stmt.Step()
	var body string
	stmt.Scan(&body)
	return body
}