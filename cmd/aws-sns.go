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

/* The sns features allow you to create a AWS SNS subscriptions url like https://<your-nsre-domain>:<nsre port>/sns/{from}/{to}/{action}
where action could be 'sms' (for sending sms) or 'call' to send voice message.

To use it you need to insert two key/val pairs into the table appconfig

insert into appconfig(key, val) values("sns_username","your sns username");
insert into appconfig(key, val) values("sns_password","your sns password");

restart nsre service after that. And then in aws SNS create a subscriptions using https protocol and url is

https://<sns_username>:<sns_password>@<your-nsre-domain>:<nsre port>/sns/{from}/{to}/{action}

replace {from} with the number you use as From number (check your twilio account)
{to} - the destination number you want to send sms/voice message
{action} - sms|call

After that you need to search log using the web browser - get into https://<your-nsre-domain>:<nsre port>/searchlog

Search using keyword 'SubscriptionConfirmation'. You will find the message from aws to find out the verify URL. Copy and paste it to the browser where you are already logged into the aws console. This will verify the SNS subscription.

To test you can publish a message in SNS and see it rings or sms your phone.

*/

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